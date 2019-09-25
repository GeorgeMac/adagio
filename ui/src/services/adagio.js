import Swagger from 'swagger-client';
import spec from '../../../pkg/rpc/controlplane/service.swagger.json';

var url = "http://localhost:7891"
if (process.env.NODE_ENV == "production") {
  // in production scenario we can use caddy proxy
  // to reach gateway api
  url = process.env.ADAGIO_API_ADDRESS || "http://localhost:8080"
}

export const Adagio = Swagger({url: url, spec: spec});
