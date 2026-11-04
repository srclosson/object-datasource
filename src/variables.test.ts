import { TemplateSrv } from '@grafana/runtime';
import { collectVariables, findLink, variableNames } from './variables';

// Resolves ${name} from a fixed set of values, calling the format function like Grafana does
const templateSrv = (values: Record<string, string | string[]>) =>
  ({
    replace: (target: string, _scoped: unknown, format: (v: unknown) => string) =>
      target.replace(/\$\{(\w+)\}/g, (match, name) => (name in values ? format(values[name]) : match)),
  }) as unknown as TemplateSrv;

describe('collectVariables', () => {
  const link = {
    name: 'a',
    query: { refId: '', expr: 'up{job=~"$job", env="${env:raw}", dc="[[dc]]"} [$__rate_interval]' } as any,
  };

  it('collects the values of the variables the link uses', () => {
    const srv = templateSrv({ job: ['api', 'web'], env: 'prod', dc: 'east', unused: 'x' });
    expect(collectVariables(link, srv, {})).toEqual({ job: ['api', 'web'], env: 'prod', dc: 'east' });
  });

  it('leaves out names that are not variables', () => {
    expect(collectVariables(link, templateSrv({}), {})).toEqual({});
  });
});

describe('findLink', () => {
  const links = [
    { id: '1', name: 'a.b', query: { refId: '' } },
    { name: 'a.c', query: { refId: '' } },
  ];

  it('finds links by id, then by name', () => {
    expect(findLink(links, { refId: 'A', linkId: '1', name: 'renamed' })).toBe(links[0]);
    expect(findLink(links, { refId: 'A', name: 'a.c' })).toBe(links[1]);
    expect(findLink(links, { refId: 'A', linkId: 'gone' })).toBeUndefined();
  });
});

describe('variableNames', () => {
  it('lists dashboard variables but not built-ins', () => {
    expect(variableNames({ expr: 'rate(up{job=~"$job", env="${env:raw}"}[$__rate_interval]) [[dc]]' })).toEqual([
      'job',
      'env',
      'dc',
    ]);
  });
});
