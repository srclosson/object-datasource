import { DataQuery, DataSourceJsonData } from '@grafana/data';

export interface ObjectQuery extends DataQuery {
  queryLink: QueryLinkConfig;
}

/**
 * These are options configured for each DataSource instance
 */
export interface ObjectDataSourceOptions extends DataSourceJsonData {
  path?: string;
  queryLinks?: QueryLinkConfig[];
}

/**
 * Value that is used in the backend, but never sent over HTTP to the frontend
 */
export interface ObjectSecureJsonData {
  basicAuthPassword?: string;
}

export type QueryLinkConfig = {
  name: string;
  query: any;
  datasourceUid: string;
};
