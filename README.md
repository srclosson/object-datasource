# Object datasource
## Vision
1. Have you ever had users that didn't know the details about how to make a query for a datasource, but needed information from that datasource?
2. Do you just want to provide your users a drop down showing a model? Much like a folder structure? 

## What it does

The object datasource tries to solve the problem that query writers sometimes don't understand the data, and query consumers sometimes don't understand the query language. 

The object datasource allows query writers to write queries, and place them in a heirarechy that data consumers can use. 



At any level, you can also decorate your data with more information. For example, perhaps you want to decorate the data for a particular `Make` with the number of active alerts for that equipment type? The object datasource should be able to provide the ability to associate a query for that data, with a level in the tree. Nodes further down in the tree, retain the meta data from queries at a higher level. 

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
