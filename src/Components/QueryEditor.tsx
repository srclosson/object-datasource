import React, { PureComponent } from 'react';
import { QueryEditorProps } from '@grafana/data';
import { DataSource } from '../datasource';
import { ObjectDataSourceOptions, ObjectQuery, QueryLinkConfig } from '../types';
import { Cascader } from 'ui-enterprise';
import { CascaderOption } from '@grafana/ui';

type Props = QueryEditorProps<DataSource, ObjectQuery, ObjectDataSourceOptions>;

export class QueryEditor extends PureComponent<Props> {
  render() {
    const { datasource, onChange, query } = this.props;
    console.log('query', query);

    if (!query.refId) {
      return <></>;
    }
    return (
      <div className="gf-form">
        <Cascader
          initialValue={query.queryLink?.name || ''}
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
          onSelect={(e: any) => {
            const queryLink: QueryLinkConfig = e as QueryLinkConfig;
            console.log('selected', queryLink);
            onChange({ ...query, queryLink });
          }}
        />
      </div>
    );
  }
}
