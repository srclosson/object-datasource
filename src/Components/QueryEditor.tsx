import React, { useMemo } from 'react';
import { css } from '@emotion/css';
import { GrafanaTheme2, QueryEditorProps } from '@grafana/data';
import { Cascader, Stack, TagList, TextLink, useStyles2 } from '@grafana/ui';
import { DataSource } from '../datasource';
import { ObjectDataSourceOptions, ObjectQuery } from '../types';
import { buildTree } from '../tree';
import { metadataFor } from '../metadata';
import { findLink } from '../variables';

const getStyles = (theme: GrafanaTheme2) => ({
  full: css`
    width: 100% !important;
  `,
  about: css`
    margin-top: ${theme.spacing(1)};
    color: ${theme.colors.text.secondary};
  `,
});

type Props = QueryEditorProps<DataSource, ObjectQuery, ObjectDataSourceOptions>;

export const QueryEditor = ({ datasource, onChange, onRunQuery, query }: Props) => {
  const styles = useStyles2(getStyles);
  const { queryLinks: links, folders } = datasource.instanceSettings.jsonData;
  const options = useMemo(() => buildTree(links ?? []), [links]);
  const link = findLink(links, query);
  const md = link ? metadataFor(link, folders) : undefined;
  const labels = Object.entries(md?.labels ?? {}).map(([k, v]) => `${k}=${v}`);

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
          const selected = links?.find((l) => l.name === path);
          if (selected) {
            onChange({ ...query, linkId: selected.id, name: selected.name });
            onRunQuery();
          }
        }}
      />
      {md && (md.description || labels.length > 0 || md.links?.length) && (
        <div className={styles.about}>
          {md.description && <div>{md.description}</div>}
          {labels.length > 0 && <TagList tags={labels} />}
          {md.links && md.links.length > 0 && (
            <Stack gap={2}>
              {md.links
                .filter((l) => l.url)
                .map((l, i) => (
                  <TextLink key={i} href={l.url} external>
                    {l.title || l.url}
                  </TextLink>
                ))}
            </Stack>
          )}
        </div>
      )}
    </div>
  );
};
