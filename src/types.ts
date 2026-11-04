import { DataSourceJsonData } from '@grafana/data';
import { DataQuery } from '@grafana/schema';

/**
 * A panel's or alert rule's query: a reference to a query link, which the backend resolves
 * from the datasource settings each time it runs
 */
export interface ObjectQuery extends DataQuery {
  linkId?: string;
  /** The link's path in the tree, for display and for links saved before they had ids */
  name?: string;
  /** Values of the dashboard variables the link's query uses, filled in when the query runs */
  variables?: Record<string, string | string[]>;
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
  serviceAccountToken?: string;
}

export interface QueryLink {
  id?: string;
  name?: string;
  uid?: string;
  query: DataQuery;
}

export const newLinkId = (): string =>
  typeof crypto !== 'undefined' && 'randomUUID' in crypto
    ? crypto.randomUUID()
    : `${Date.now().toString(36)}-${Math.random().toString(36).slice(2, 10)}`;
