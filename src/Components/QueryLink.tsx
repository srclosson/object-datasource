import React, { useEffect } from 'react';
import { css } from '@emotion/css';
import { DataSourceInstanceSettings, GrafanaTheme2 } from '@grafana/data';
import { Button, InlineField, Input, useStyles2 } from '@grafana/ui';
import { DataSourcePicker, getDataSourceSrv } from '@grafana/runtime';
import { Metadata, QueryLink } from '../types';
import { metadataFor } from '../metadata';
import { DatasourceQueryEditor } from './DatasourceQueryEditor';
import { MetadataEditor } from './MetadataEditor';
import { Preview } from './Preview';

const getStyles = (theme: GrafanaTheme2) => ({
  row: css`
    display: flex;
    gap: ${theme.spacing(1)};
    align-items: flex-start;
  `,
  editor: css`
    display: flex;
    flex-direction: column;
    padding: ${theme.spacing(1, 0)};
  `,
});

type Props = {
  value: QueryLink;
  folders?: Record<string, Metadata>;
  onChange: (value: QueryLink) => void;
  onDelete: () => void;
  className?: string;
};

export const QueryLinkEditor = ({ value, folders, onChange, onDelete, className }: Props) => {
  const styles = useStyles2(getStyles);

  // A new link starts on the default datasource
  useEffect(() => {
    if (!value.uid) {
      const ds = getDataSourceSrv().getInstanceSettings(null);
      if (ds) {
        onChange({ ...value, uid: ds.uid, query: { refId: '', datasource: { uid: ds.uid, type: ds.type } } });
      }
    }
    // Only for links without a datasource; onChange and value change on every edit
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [value.uid]);

  // What the link inherits: the metadata of the folders above it, without its own
  const inherited = metadataFor({ ...value, metadata: undefined }, folders);

  return (
    <div className={className}>
      <div className={styles.row}>
        <InlineField
          label="Name"
          labelWidth={12}
          grow
          tooltip="Where the link appears in the tree. Dots separate folders, e.g. plant.line1.temperature"
        >
          <Input
            value={value.name ?? ''}
            placeholder="folder.subfolder.name"
            onChange={(e) => onChange({ ...value, name: e.currentTarget.value })}
          />
        </InlineField>
        <Button
          variant="destructive"
          title="Remove query link"
          icon="times"
          onClick={(event) => {
            event.preventDefault();
            onDelete();
          }}
        />
      </div>
      <InlineField label="Datasource" labelWidth={12}>
        <DataSourcePicker
          current={value.uid}
          onChange={(ds: DataSourceInstanceSettings) =>
            onChange({
              ...value,
              uid: ds.uid,
              query: { ...value.query, datasource: { uid: ds.uid, type: ds.type } },
            })
          }
        />
      </InlineField>
      <div className={styles.editor}>
        <DatasourceQueryEditor uid={value.uid} query={value.query} onChange={(query) => onChange({ ...value, query })} />
        <Preview uid={value.uid} query={value.query} mode="data" />
      </div>
      <MetadataEditor
        label="Metadata"
        value={value.metadata}
        inherited={inherited}
        onChange={(metadata) => onChange({ ...value, metadata })}
      />
    </div>
  );
};
