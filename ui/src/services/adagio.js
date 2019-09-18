import Swagger from 'swagger-client';
import spec from '../../../pkg/rpc/controlplane/service.swagger.json';

export const Adagio = Swagger({url: "http://localhost:7891", spec: spec});
