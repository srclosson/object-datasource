import { buildTree, groupLinks, savedNames } from './tree';

const link = (name?: string) => ({ name, query: { refId: '' } });

describe('buildTree', () => {
  it('nests links by their dot-separated names', () => {
    expect(buildTree([link('plant.line1.temp'), link('plant.line2.temp')])).toEqual([
      {
        label: 'plant',
        value: 'plant',
        items: [
          { label: 'line1', value: 'plant.line1', items: [{ label: 'temp', value: 'plant.line1.temp' }] },
          { label: 'line2', value: 'plant.line2', items: [{ label: 'temp', value: 'plant.line2.temp' }] },
        ],
      },
    ]);
  });

  it('lets a link also be a folder, in either order', () => {
    const expected = [
      { label: 'a', value: 'a', items: [{ label: 'b', value: 'a.b', items: [{ label: 'c', value: 'a.b.c' }] }] },
    ];
    expect(buildTree([link('a.b'), link('a.b.c')])).toEqual(expected);
    expect(buildTree([link('a.b.c'), link('a.b')])).toEqual(expected);
  });

  it('leaves out links without a name', () => {
    expect(buildTree([link(), link('')])).toEqual([]);
  });
});

describe('groupLinks', () => {
  const links = [
    { id: '1', name: 'plant.line1.temp', query: { refId: '' } },
    { id: '2', name: 'plant.line1.pressure', query: { refId: '' } },
    { id: '3', name: 'plant.power', query: { refId: '' } },
    { id: '4', name: 'top', query: { refId: '' } },
    { id: '5', name: 'plant.new', query: { refId: '' } },
  ];

  it('keeps unsaved links apart and folders saved ones by their saved names', () => {
    // Link 3 was renamed since the save; link 5 is new
    const saved = savedNames([...links.slice(0, 2), { ...links[2], name: 'plant.line2.power' }, links[3]]);
    const { fresh, root } = groupLinks(links, saved);

    expect(fresh).toEqual([4]);
    expect(root.links).toEqual([3]);
    expect(root.folders.map((f) => f.path)).toEqual(['plant']);
    const plant = root.folders[0];
    expect(plant.links).toEqual([]);
    expect(plant.folders.map((f) => [f.path, f.links])).toEqual([
      ['plant.line1', [0, 1]],
      ['plant.line2', [2]],
    ]);
  });

  it('knows links without ids by name', () => {
    const legacy = [{ name: 'a.b', query: { refId: '' } }];
    expect(groupLinks(legacy, savedNames(legacy)).fresh).toEqual([]);
  });
});
