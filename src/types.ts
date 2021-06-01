import { DataQuery, DataSourceJsonData } from '@grafana/data';

export interface ObjectQuery extends DataQuery {
  name: string;
  config: QueryLinkConfig;
}

/**
 * These are options configured for each DataSource instance
 */
export interface ObjectDataSourceOptions extends DataSourceJsonData {
  queryLinks?: QueryLinkConfig[];
}

/**
 * Value that is used in the backend, but never sent over HTTP to the frontend
 */
export interface ObjectSecureJsonData {
  basicAuthPassword?: string;
}

export interface QueryLink {
  name?: string;
  uid?: string;
}

export interface ConfigDataQuery extends DataQuery {
  datasourceId: number;
}

export interface QueryLinkConfig extends QueryLink {
  query?: ConfigDataQuery;
}

export interface ObjectPanelType {}
