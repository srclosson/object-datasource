import React, { useMemo } from 'react';
import { css } from '@emotion/css';
import { QueryEditorProps } from '@grafana/data';
import { Cascader } from '@grafana/ui';
import { DataSource } from '../datasource';
import { ObjectDataSourceOptions, ObjectQuery } from '../types';
import { buildTree } from '../tree';

const styles = {
  full: css`
    width: 100% !important;
  `,
};

type Props = QueryEditorProps<DataSource, ObjectQuery, ObjectDataSourceOptions>;

export const QueryEditor = ({ datasource, onChange, onRunQuery, query }: Props) => {
  const links = datasource.instanceSettings.jsonData.queryLinks;
  const options = useMemo(() => buildTree(links ?? []), [links]);

  return (
    <div className={styles.full}>
      <Cascader
        initialValue={query.name}
        separator="."
        changeOnSelect={true}
        displayAllSelectedLevels={true}
        options={options}
        onSelect={(path: string) => {
          // Folders can be selected too; only a path that is a link changes the query
          const link = links?.find((l) => l.name === path);
          if (link) {
            onChange({ ...query, linkId: link.id, name: link.name });
            onRunQuery();
          }
        }}
      />
    </div>
  );
};
