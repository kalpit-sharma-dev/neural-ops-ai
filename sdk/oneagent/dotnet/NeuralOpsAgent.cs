using System.Diagnostics;
using System.Net.Http.Headers;
using System.Text;
using System.Text.Json;

namespace NeuralOps.OneAgent;

/// <summary>HTTP auto-instrumentation handler for .NET services.</summary>
public sealed class NeuralOpsHandler : DelegatingHandler
{
    private readonly NeuralOpsConfig _config;

    public NeuralOpsHandler(NeuralOpsConfig config, HttpMessageHandler? inner = null)
        : base(inner ?? new HttpClientHandler())
    {
        _config = config;
    }

    protected override async Task<HttpResponseMessage> SendAsync(HttpRequestMessage request, CancellationToken cancellationToken)
    {
        var sw = Stopwatch.StartNew();
        var response = await base.SendAsync(request, cancellationToken);
        sw.Stop();
        _ = EmitSpanAsync(request, response, sw.ElapsedMilliseconds);
        return response;
    }

    private async Task EmitSpanAsync(HttpRequestMessage request, HttpResponseMessage response, long durationMs)
    {
        try
        {
            using var client = new HttpClient();
            var payload = JsonSerializer.Serialize(new
            {
                functionName = $"{request.Method} {request.RequestUri?.AbsolutePath}",
                filePath = "System.Net.Http",
                lineNo = (int)response.StatusCode,
                selfTimeMs = durationMs,
                sampleCount = 1,
                service = _config.ServiceName,
            });
            using var content = new StringContent(payload, Encoding.UTF8, "application/json");
            content.Headers.ContentType = new MediaTypeHeaderValue("application/json");
            var req = new HttpRequestMessage(HttpMethod.Post, _config.IngestUrl) { Content = content };
            req.Headers.Add("X-Tenant-ID", _config.TenantId);
            await client.SendAsync(req);
        }
        catch
        {
            // best-effort
        }
    }
}

public sealed class NeuralOpsConfig
{
    public string ServiceName { get; init; } = Environment.GetEnvironmentVariable("NEURALOPS_SERVICE") ?? "dotnet-service";
    public string IngestUrl { get; init; } = Environment.GetEnvironmentVariable("NEURALOPS_INGEST_URL") ?? "http://localhost:8080/api/v1/apm/profiles";
    public string TenantId { get; init; } = Environment.GetEnvironmentVariable("NEURALOPS_TENANT_ID") ?? "default";
}
