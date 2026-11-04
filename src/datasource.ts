import { DataSourceInstanceSettings } from '@grafana/data';
import { DataSourceWithBackend } from '@grafana/runtime';
import { ObjectDataSourceOptions, ObjectQuery } from './types';

// Queries run in the backend, which resolves each query's link from the datasource settings,
// so panels and alert rules behave the same.
export class DataSource extends DataSourceWithBackend<ObjectQuery, ObjectDataSourceOptions> {
  constructor(public instanceSettings: DataSourceInstanceSettings<ObjectDataSourceOptions>) {
    super(instanceSettings);
  }

  filterQuery(query: ObjectQuery): boolean {
    return !query.hide && Boolean(query.linkId || query.name);
  }
}
