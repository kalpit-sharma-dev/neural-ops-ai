FROM node:20-alpine
WORKDIR /agent
COPY index.js package.json ./
ENV NEURALOPS_SERVICE=node-service
ENV NEURALOPS_INGEST_URL=http://gateway:8080/api/v1/apm/profiles
ENV NEURALOPS_TENANT_ID=default
CMD ["node", "-e", "const {init}=require('./index'); init(); setInterval(()=>{}, 60000)"]
