import { CascaderOption } from '@grafana/ui';
import { QueryLink } from './types';

/**
 * Builds the tree the query editor shows from the links' dot-separated names. Each segment is
 * a node whose value is its path, so a name can be both a link and a folder of other links
 * (`plant.line1` and `plant.line1.temp`). Links without a name can't be placed and are left out.
 */
export const buildTree = (links: QueryLink[]): CascaderOption[] => {
  const tree: CascaderOption[] = [];
  for (const link of links) {
    if (!link.name) {
      continue;
    }
    const segments = link.name.split('.');
    let level = tree;
    segments.forEach((label, i) => {
      const value = segments.slice(0, i + 1).join('.');
      let node = level.find((n) => n.label === label);
      if (!node) {
        node = { label, value };
        level.push(node);
      }
      if (i < segments.length - 1) {
        node.items = node.items ?? [];
        level = node.items;
      }
    });
  }
  return tree;
};

/** A folder in the config editor: its path, sub-folders and the indexes of its links */
export interface FolderNode {
  path: string;
  label: string;
  folders: FolderNode[];
  links: number[];
}

/** Identifies a link across edits; links saved before they had ids are known by name */
export const linkKey = (link: QueryLink): string => link.id ?? `name:${link.name ?? ''}`;

/** The name each link had when the settings were last saved */
export const savedNames = (links: QueryLink[] = []): Map<string, string> =>
  new Map(links.map((l) => [linkKey(l), l.name ?? '']));

/**
 * Arranges the config editor's links: links added since the last save stay apart, and saved
 * links go into folders by the name they were saved with, so renaming one doesn't move it
 * until the next save.
 */
export const groupLinks = (links: QueryLink[], saved: Map<string, string>): { fresh: number[]; root: FolderNode } => {
  const root: FolderNode = { path: '', label: '', folders: [], links: [] };
  const fresh: number[] = [];
  links.forEach((link, index) => {
    // A link saved before links had ids gets one when first edited; it was saved by name
    const name = saved.get(linkKey(link)) ?? saved.get(`name:${link.name ?? ''}`);
    if (name === undefined) {
      fresh.push(index);
      return;
    }
    const segments = name.split('.');
    let folder = root;
    segments.slice(0, -1).forEach((label, i) => {
      const path = segments.slice(0, i + 1).join('.');
      let child = folder.folders.find((f) => f.label === label);
      if (!child) {
        child = { path, label, folders: [], links: [] };
        folder.folders.push(child);
      }
      folder = child;
    });
    folder.links.push(index);
  });
  return { fresh, root };
};
