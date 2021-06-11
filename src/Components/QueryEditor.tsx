import React, { PureComponent } from 'react';
import { QueryEditorProps } from '@grafana/data';
import { DataSource } from '../datasource';
import { ObjectDataSourceOptions, ObjectQuery, QueryLink } from '../types';
import { CascaderOption, Cascader } from '@grafana/ui';
import { css } from 'emotion';

const styles = {
  full: css`
    width: 100% !important;
  `,
};

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
            const existingNodeIndex = node.findIndex((e) => e.label === splitName[i]);

            if (existingNodeIndex >= 0) {
              node = node[existingNodeIndex].items!;
            } else {
              const newNode: CascaderOption = {
                label: splitName[i],
                value: [...splitName.slice(0, i), splitName[i]].join('.'), //q,
              };
              node.push(newNode);
              if (i < splitName.length - 1) {
                newNode.items = [];
                node = newNode.items;
              }
            }
          }
          node = tree;
        }
      });
    }
    return tree;
  };

  render() {
    const { datasource, onChange, onRunQuery, query } = this.props;
    console.log(query);
    return (
      <div className={styles.full}>
        <Cascader
          initialValue={query.name}
          separator="."
          changeOnSelect={true}
          displayAllSelectedLevels={true}
          options={this.buildTree()}
          onSelect={(e: string) => {
            const queryLink: QueryLink | undefined = datasource.instanceSettings.jsonData.queryLinks?.find(
              (ql: QueryLink) => ql.name && e === ql.name
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
