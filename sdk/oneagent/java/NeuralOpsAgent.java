package com.neuralops.oneagent;

import okhttp3.Interceptor;
import okhttp3.OkHttpClient;
import okhttp3.Request;
import okhttp3.Response;
import org.json.JSONObject;

import java.io.IOException;
import java.util.concurrent.TimeUnit;

/** Lightweight HTTP auto-instrumentation for Java services. */
public final class NeuralOpsAgent {
    private NeuralOpsAgent() {}

    public static OkHttpClient.Builder instrument(OkHttpClient.Builder builder, Config config) {
        return builder.addInterceptor(new SpanInterceptor(config));
    }

    public static final class Config {
        public String serviceName = System.getenv().getOrDefault("NEURALOPS_SERVICE", "java-service");
        public String ingestUrl = System.getenv().getOrDefault("NEURALOPS_INGEST_URL", "http://localhost:8080/api/v1/apm/profiles");
        public String tenantId = System.getenv().getOrDefault("NEURALOPS_TENANT_ID", "default");
    }

    static final class SpanInterceptor implements Interceptor {
        private final Config config;
        SpanInterceptor(Config config) { this.config = config; }

        @Override
        public Response intercept(Chain chain) throws IOException {
            Request request = chain.request();
            long start = System.nanoTime();
            Response response = chain.proceed(request);
            try {
                JSONObject body = new JSONObject();
                body.put("functionName", request.method() + " " + request.url().encodedPath());
                body.put("filePath", "okhttp3");
                body.put("lineNo", response.code());
                body.put("selfTimeMs", TimeUnit.NANOSECONDS.toMillis(System.nanoTime() - start));
                body.put("sampleCount", 1);
                body.put("service", config.serviceName);
                okhttp3.Request emit = new okhttp3.Request.Builder()
                    .url(config.ingestUrl)
                    .post(okhttp3.RequestBody.create(body.toString(), okhttp3.MediaType.parse("application/json")))
                    .header("X-Tenant-ID", config.tenantId)
                    .build();
                new OkHttpClient().newCall(emit).execute().close();
            } catch (Exception ignored) { }
            return response;
        }
    }
}
