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
