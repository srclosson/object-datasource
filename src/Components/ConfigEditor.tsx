import React, { PureComponent } from 'react';
import { DataSourceHttpSettings } from '@grafana/ui';
import { DataSourcePluginOptionsEditorProps, DataSourceSettings } from '@grafana/data';
import { ObjectDataSourceOptions, ObjectSecureJsonData } from '../types';
import { QueryLinks } from './QueryLinks';

interface Props extends DataSourcePluginOptionsEditorProps<ObjectDataSourceOptions, ObjectSecureJsonData> {}

interface State {}

export class ConfigEditor extends PureComponent<Props, State> {
  render() {
    const { options, onOptionsChange } = this.props;
    const { jsonData } = options;

    console.log('jsonData', jsonData, options);
    return (
      <div className="gf-form-group">
        <div className="gf-form">
          <DataSourceHttpSettings
            defaultUrl="http://localhost:3000"
            dataSourceConfig={options}
            showAccessOptions={true}
            onChange={(val: DataSourceSettings<ObjectDataSourceOptions, ObjectSecureJsonData>) => {
              console.log('we got updated settings', val);
              onOptionsChange(val);
            }}
          />
        </div>
        <QueryLinks
          value={jsonData.queryLinks || []}
          onChange={(e) => {
            console.log('we got e/querylinks', e);
            onOptionsChange({
              ...options,
              jsonData: {
                ...options.jsonData,
                queryLinks: e,
              },
            });
          }}
        />
      </div>
    );
  }
}
