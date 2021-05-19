import { DataSourceInstanceSettings, DataQueryRequest, DataQueryResponse } from '@grafana/data';
import { DataSourceWithBackend, HealthCheckResult } from '@grafana/runtime';
import { ObjectDataSourceOptions, ObjectQuery } from './types';
import { Observable } from 'rxjs';

export class DataSource extends DataSourceWithBackend<ObjectQuery, ObjectDataSourceOptions> {
  constructor(public instanceSettings: DataSourceInstanceSettings<ObjectDataSourceOptions>) {
    super(instanceSettings);
  }

  query(request: DataQueryRequest<ObjectQuery>): Observable<DataQueryResponse> {
    console.log('request', request, this.instanceSettings);
    return super.query(request);
  }

  callHealthCheck(): Promise<HealthCheckResult> {
    console.log('calling callHealthCheck');
    return super.callHealthCheck();
  }
}
