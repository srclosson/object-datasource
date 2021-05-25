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
                  value: q.name,
                  label: q.name,
                } as CascaderOption)
            ) || []
          }
          onSelect={(e: string) => {
            console.log('e', e);
            const queryLink: QueryLinkConfig | undefined = datasource.instanceSettings.jsonData.queryLinks?.find(
              (ql: QueryLinkConfig) => ql.name && e === ql.name
            );
            if (queryLink && queryLink.query && queryLink.name) {
              console.log('changing. Using queryLink', queryLink);
              onChange({ ...query, name: queryLink.name, config: queryLink });
            }
          }}
        />
      </div>
    );
  }
}
