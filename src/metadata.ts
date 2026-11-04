import { DataFrame, FieldType } from '@grafana/data';
import { Metadata, QueryLink } from './types';

/**
 * `child` on top of `parent`: the child's own values win, labels merge with the child's
 * winning, and links and label lookups accumulate, parent's first. Matches the backend.
 */
export const inherit = (child: Metadata | undefined, parent: Metadata): Metadata => {
  if (!child) {
    return parent;
  }
  const out: Metadata = { ...parent };
  for (const key of ['description', 'unit', 'displayName'] as const) {
    if (child[key]) {
      out[key] = child[key];
    }
  }
  for (const key of ['min', 'max', 'cacheTtl'] as const) {
    if (child[key] !== undefined) {
      out[key] = child[key];
    }
  }
  if (child.thresholds?.length) {
    out.thresholds = child.thresholds;
  }
  if (parent.labels || child.labels) {
    out.labels = { ...parent.labels, ...child.labels };
  }
  if (child.links?.length) {
    out.links = [...(parent.links ?? []), ...child.links];
  }
  if (child.labelLookups?.length) {
    out.labelLookups = [...(parent.labelLookups ?? []), ...child.labelLookups];
  }
  return out;
};

/** A link's metadata on top of each folder above it, outermost first */
export const metadataFor = (link: QueryLink, folders: Record<string, Metadata> = {}): Metadata => {
  const segments = (link.name ?? '').split('.');
  let md: Metadata = {};
  for (let i = 1; i < segments.length; i++) {
    md = inherit(folders[segments.slice(0, i).join('.')], md);
  }
  return inherit(link.metadata, md);
};

/** Every folder above the links, in tree order */
export const folderPaths = (links: QueryLink[]): string[] => {
  const paths = new Set<string>();
  for (const link of links) {
    const segments = (link.name ?? '').split('.');
    for (let i = 1; i < segments.length; i++) {
      paths.add(segments.slice(0, i).join('.'));
    }
  }
  return [...paths].sort();
};

/**
 * Query results as rows of labels, the way a label lookup reads them: a table's columns, or a
 * series' labels. Matches the backend.
 */
export const labelRows = (frames: DataFrame[]): Array<Record<string, string>> => {
  const rows: Array<Record<string, string>> = [];
  for (const frame of frames) {
    for (let i = 0; i < frame.length; i++) {
      const row: Record<string, string> = {};
      for (const field of frame.fields) {
        if (field.type === FieldType.time) {
          continue;
        }
        if (field.labels && Object.keys(field.labels).length) {
          Object.assign(row, field.labels);
          continue;
        }
        const value = field.values[i];
        if (value !== null && value !== undefined) {
          row[field.name] = String(value);
        }
      }
      if (Object.keys(row).length) {
        rows.push(row);
      }
    }
  }
  return rows;
};
