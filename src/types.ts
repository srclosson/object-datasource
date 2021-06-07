import { DataQuery, DataSourceJsonData } from '@grafana/data';

export interface ObjectQuery extends DataQuery {
  name: string;
  config: QueryLink;
}

/**
 * These are options configured for each DataSource instance
 */
export interface ObjectDataSourceOptions extends DataSourceJsonData {
  queryLinks?: QueryLink[];
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
  query: DataQuery;
}

export interface ObjectPanelType {}
