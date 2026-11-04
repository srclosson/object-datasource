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
  /** Metadata of folders in the tree, by path; links inherit it */
  folders?: Record<string, Metadata>;
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
  metadata?: Metadata;
}

/** Decorates the data under a folder or query link */
export interface Metadata {
  description?: string;
  unit?: string;
  displayName?: string;
  labels?: Record<string, string>;
  /** References such as runbooks and manufacturers' pages, clickable on the data */
  links?: Reference[];
  min?: number;
  max?: number;
  /** Levels above which the data needs attention, lowest first */
  thresholds?: Threshold[];
  /** Queries against other datasources whose results become labels */
  labelLookups?: LabelLookup[];
}

export interface Reference {
  title: string;
  url: string;
}

export interface Threshold {
  value: number;
  color: string;
}

/**
 * A query whose results become labels: a table's columns, or a series' labels. With
 * matchLabel, each series gets the row whose matchColumn (default: the same name) equals the
 * series' matchLabel; without it, every series gets the first row.
 */
export interface LabelLookup {
  uid?: string;
  query: DataQuery;
  matchLabel?: string;
  matchColumn?: string;
}

export const newLinkId = (): string =>
  typeof crypto !== 'undefined' && 'randomUUID' in crypto
    ? crypto.randomUUID()
    : `${Date.now().toString(36)}-${Math.random().toString(36).slice(2, 10)}`;
