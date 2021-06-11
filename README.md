# Object datasource
## Vision
1. Have you ever had users that didn't know the details about how to make a query for a datasource, but needed information from that datasource?
2. Do you just want to provide your users a drop down showing a model? Much like a folder structure? 

## What does it do?
The object datasource tries to solve the problem between query creators (those knowledgable about the a query language) and query consumers (those knowledgable about the data a query produces)
Ideally, everyone interacting with data is knowledgable as both a query creator, and query consumer, but this is often not the case.

The object datasource allows query creators to create meaningful queryies, and associate them in a heiarchy that can be accessed by query consumers. 
Query consumers (those who are knowledgable about how to use and interpret the data) then have much easier access to the data they need to be successful.

## What does it look like?
Query creators write queries in the datasource config like so:


Query consumers access the data through a tree like heirarchy like so:


### Frontend

1. Install dependencies

   ```bash
   yarn install
   ```

2. Build plugin in development mode or run in watch mode

   ```bash
   yarn dev
   ```

   or

   ```bash
   yarn watch
   ```

3. Build plugin in production mode

   ```bash
   yarn build
   ```

### Backend

1. Update [Grafana plugin SDK for Go](https://grafana.com/docs/grafana/latest/developers/plugins/backend/grafana-plugin-sdk-for-go/) dependency to the latest minor version:

   ```bash
   go get -u github.com/grafana/grafana-plugin-sdk-go
   ```

2. Build backend plugin binaries for Linux, Windows and Darwin:

   ```bash
   mage -v
   ```

3. List all available Mage targets for additional commands:

   ```bash
   mage -l
   ```

## Learn more

- [Build a data source backend plugin tutorial](https://grafana.com/tutorials/build-a-data-source-backend-plugin)
- [Grafana documentation](https://grafana.com/docs/)
- [Grafana Tutorials](https://grafana.com/tutorials/) - Grafana Tutorials are step-by-step guides that help you make the most of Grafana
- [Grafana UI Library](https://developers.grafana.com/ui) - UI components to help you build interfaces using Grafana Design System
- [Grafana plugin SDK for Go](https://grafana.com/docs/grafana/latest/developers/plugins/backend/grafana-plugin-sdk-for-go/)
