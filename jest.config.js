// force timezone to UTC to allow tests to work regardless of local timezone
// generally used by snapshots, but can affect specific tests
process.env.TZ = 'UTC';

const base = require('./.config/jest.config');

module.exports = {
  // Jest configuration provided by Grafana scaffolding
  ...base,
  transform: {
    ...base.transform,
    // @grafana/data pulls in ES modules shipped as .mjs (d3-color), which the scaffold's pattern misses
    '^.+\\.mjs$': base.transform['^.+\\.(t|j)sx?$'],
  },
};
