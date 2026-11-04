import { FieldType, toDataFrame } from '@grafana/data';
import { folderPaths, labelRows, metadataFor } from './metadata';

describe('metadataFor', () => {
  it('inherits from the folders above the link', () => {
    const folders = {
      plant: { unit: 'celsius', labels: { site: 'plant1', owner: 'ops' }, links: [{ title: 'Site', url: 'https://s' }] },
      'plant.line1': { labels: { owner: 'line1' }, links: [{ title: 'Line', url: 'https://l' }] },
    };
    const link = { name: 'plant.line1.temp', query: { refId: '' }, metadata: { description: 'Temp', max: 120 } };

    expect(metadataFor(link, folders)).toEqual({
      unit: 'celsius',
      description: 'Temp',
      max: 120,
      labels: { site: 'plant1', owner: 'line1' },
      links: [
        { title: 'Site', url: 'https://s' },
        { title: 'Line', url: 'https://l' },
      ],
    });
  });
});

describe('folderPaths', () => {
  it('lists each folder once, in order', () => {
    const links = [{ name: 'b.x.y', query: { refId: '' } }, { name: 'a.z', query: { refId: '' } }, { name: 'top', query: { refId: '' } }];
    expect(folderPaths(links)).toEqual(['a', 'b', 'b.x']);
  });
});

describe('labelRows', () => {
  it('reads table columns and series labels, but not values or times', () => {
    const table = toDataFrame({
      fields: [
        { name: 'host', type: FieldType.string, values: ['a', 'b'] },
        { name: 'maker', type: FieldType.string, values: ['Acme', 'Globex'] },
      ],
    });
    const series = toDataFrame({
      fields: [
        { name: 'time', type: FieldType.time, values: [1] },
        { name: 'value', type: FieldType.number, values: [1], labels: { instance: 'c', machine: 'arm64' } },
      ],
    });
    expect(labelRows([table, series])).toEqual([
      { host: 'a', maker: 'Acme' },
      { host: 'b', maker: 'Globex' },
      { instance: 'c', machine: 'arm64' },
    ]);
  });
});
