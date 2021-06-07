import React, { PureComponent } from 'react';
import { QueryEditorProps } from '@grafana/data';
import { DataSource } from '../datasource';
import { ObjectDataSourceOptions, ObjectQuery, QueryLink } from '../types';
import { CascaderOption, Cascader } from '@grafana/ui';

type Props = QueryEditorProps<DataSource, ObjectQuery, ObjectDataSourceOptions>;

export class QueryEditor extends PureComponent<Props> {
  buildTree = (): CascaderOption[] => {
    const tree: CascaderOption[] = [];
    let node: CascaderOption[] = tree;
    const { queryLinks } = this.props.datasource.instanceSettings.jsonData;
    if (queryLinks) {
      queryLinks.forEach((q: QueryLink) => {
        // If the user has set the name of the query link
        if (q.name) {
          const splitName = q.name.split('.');
          for (let i = 0; i < splitName.length; i++) {
            // If the node is already here, don't add it again.
            console.log('node', node);
            const existingNodeIndex = node.findIndex((e) => e.label === splitName[i]);

            if (existingNodeIndex >= 0) {
              node = node[existingNodeIndex].items!;
            } else {
              const newItems: CascaderOption[] = [];
              node.push({
                label: splitName[i],
                value: splitName[i],
                items: newItems,
              });
              node = newItems;
            }
          }
          node = tree;
        }
      });
    }
    return tree;
  };

  getOptions = (): CascaderOption[] => {
    const { datasource } = this.props;
    return (
      datasource.instanceSettings.jsonData.queryLinks?.map(
        (q: QueryLink) =>
          ({
            value: q.name,
            label: q.name,
          } as CascaderOption)
      ) || []
    );
  };

  render() {
    const { datasource, onChange, onRunQuery, query } = this.props;
    console.log('query', query);

    return (
      <div className="gf-form">
        <Cascader
          initialValue={query.name}
          separator="."
          changeOnSelect
          displayAllSelectedLevels
          options={this.buildTree()}
          onSelect={(e: string) => {
            if (!query.name || query.name === '') {
              query.name = e;
            } else {
              query.name += '.' + e;
            }

            const queryLink: QueryLink | undefined = datasource.instanceSettings.jsonData.queryLinks?.find(
              (ql: QueryLink) => ql.name && query.name === ql.name
            );
            if (queryLink && queryLink.query && queryLink.name) {
              onChange({ ...query, name: queryLink.name, config: queryLink });
              onRunQuery();
            } else {
              onChange({ ...query, name: query.name });
            }
          }}
        />
      </div>
    );
  }
}
