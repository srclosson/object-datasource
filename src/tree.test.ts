import { buildTree } from './tree';

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
