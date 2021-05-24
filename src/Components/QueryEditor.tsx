import React, { PureComponent } from 'react';
import { QueryEditorProps } from '@grafana/data';
import { DataSource } from '../datasource';
import { ObjectDataSourceOptions, ObjectQuery, QueryLinkConfig } from '../types';
import { CascaderOption, Cascader } from '@grafana/ui';

type Props = QueryEditorProps<DataSource, ObjectQuery, ObjectDataSourceOptions>;

export class QueryEditor extends PureComponent<Props> {
  render() {
    const { datasource, onChange, query } = this.props;
    console.log('query', query);

    return (
      <div className="gf-form">
        <Cascader
          initialValue={query.name}
          changeOnSelect
          displayAllSelectedLevels
          options={
            datasource.instanceSettings.jsonData.queryLinks?.map(
              (q: QueryLinkConfig) =>
                ({
                  value: q,
                  label: q.name,
                } as CascaderOption)
            ) || []
          }
          onSelect={(e: string) => {
            console.log('e', e);
            const queryLink: QueryLinkConfig = e as QueryLinkConfig;
            console.log('selected', queryLink);
            if (queryLink.query) {
              onChange({ ...query, config: queryLink });
            }
          }}
        />
      </div>
    );
  }
}
