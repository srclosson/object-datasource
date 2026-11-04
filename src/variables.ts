import { ScopedVars } from '@grafana/data';
import { TemplateSrv } from '@grafana/runtime';
import { ObjectQuery, QueryLink } from './types';

// $name, ${name}, ${name:format}, [[name]] and [[name:format]], as Grafana writes them
const variablePattern = /\$(\w+)|\$\{(\w+)(?::\w+)?\}|\[\[(\w+)(?::\w+)?\]\]/g;

export const findLink = (links: QueryLink[] | undefined, query: ObjectQuery): QueryLink | undefined =>
  links?.find((l) => query.linkId && l.id === query.linkId) ?? links?.find((l) => query.name && l.name === query.name);

/**
 * The current values of the variables a link's query uses. The backend substitutes them into
 * the link's query; values only, so a query can't be rewritten from the browser.
 */
export const collectVariables = (
  link: QueryLink,
  templateSrv: TemplateSrv,
  scopedVars: ScopedVars
): Record<string, string | string[]> => {
  const names = new Set<string>();
  for (const m of JSON.stringify(link.query ?? {}).matchAll(variablePattern)) {
    names.add(m[1] ?? m[2] ?? m[3]);
  }

  const variables: Record<string, string | string[]> = {};
  for (const name of names) {
    let value: unknown;
    // The format function receives the resolved value (all values for "All"), and is only
    // called for names that are variables
    templateSrv.replace(`\${${name}}`, scopedVars, (v: unknown) => {
      value = v;
      return '';
    });
    if (value !== undefined) {
      variables[name] = Array.isArray(value) ? value.map(String) : String(value);
    }
  }
  return variables;
};
