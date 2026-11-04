import { DataSourceInstanceSettings, ScopedVars } from '@grafana/data';
import { DataSourceWithBackend, getTemplateSrv } from '@grafana/runtime';
import { ObjectDataSourceOptions, ObjectQuery } from './types';
import { collectVariables, findLink } from './variables';

// Queries run in the backend, which resolves each query's link from the datasource settings,
// so panels and alert rules behave the same.
export class DataSource extends DataSourceWithBackend<ObjectQuery, ObjectDataSourceOptions> {
  constructor(public instanceSettings: DataSourceInstanceSettings<ObjectDataSourceOptions>) {
    super(instanceSettings);
  }

  filterQuery(query: ObjectQuery): boolean {
    return !query.hide && Boolean(query.linkId || query.name);
  }

  applyTemplateVariables(query: ObjectQuery, scopedVars: ScopedVars): ObjectQuery {
    const link = findLink(this.instanceSettings.jsonData.queryLinks, query);
    if (!link) {
      return query;
    }
    const variables = collectVariables(link, getTemplateSrv(), scopedVars);
    return Object.keys(variables).length ? { ...query, variables } : query;
  }
}
