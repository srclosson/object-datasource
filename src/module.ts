import { DataSourcePlugin } from '@grafana/data';
import { DataSource } from './datasource';
import { ConfigEditor } from './Components/ConfigEditor';
import { QueryEditor } from './Components/QueryEditor';
import { ObjectQuery, ObjectDataSourceOptions } from './types';

export const plugin = new DataSourcePlugin<DataSource, ObjectQuery, ObjectDataSourceOptions>(DataSource)
  .setConfigEditor(ConfigEditor)
  .setQueryEditor(QueryEditor);
