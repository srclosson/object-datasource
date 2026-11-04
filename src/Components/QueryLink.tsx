import React, { useEffect, useMemo, useState } from 'react';
import { css } from '@emotion/css';
import {
  DataSourceApi,
  DataSourceInstanceSettings,
  GrafanaTheme2,
  LoadingState,
  PanelData,
  getDefaultTimeRange,
} from '@grafana/data';
import { DataQuery } from '@grafana/schema';
import { Button, InlineField, Input, useStyles2 } from '@grafana/ui';
import { DataSourcePicker, getDataSourceSrv } from '@grafana/runtime';
import { QueryLink } from '../types';

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
  onChange: (value: QueryLink) => void;
  onDelete: () => void;
  className?: string;
};

export const QueryLinkEditor = ({ value, onChange, onDelete, className }: Props) => {
  const styles = useStyles2(getStyles);
  const [datasource, setDatasource] = useState<DataSourceApi>();

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

  useEffect(() => {
    let cancelled = false;
    if (value.uid) {
      getDataSourceSrv()
        .get(value.uid)
        .then((ds) => !cancelled && setDatasource(ds))
        .catch(() => !cancelled && setDatasource(undefined));
    }
    return () => {
      cancelled = true;
    };
  }, [value.uid]);

  const data: PanelData = useMemo(
    () => ({ state: LoadingState.NotStarted, series: [], timeRange: getDefaultTimeRange() }),
    []
  );

  const renderQueryEditor = () => {
    if (!datasource) {
      return null;
    }
    const QueryEditor = datasource.components?.QueryEditor;
    if (!QueryEditor) {
      return <div>Data source plugin does not export any Query Editor component</div>;
    }
    return (
      <QueryEditor
        key={datasource.uid}
        query={value.query}
        datasource={datasource}
        onChange={(query: DataQuery) => onChange({ ...value, query })}
        onRunQuery={() => {}}
        data={data}
        range={data.timeRange}
        queries={[value.query]}
      />
    );
  };

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
      <div className={styles.editor}>{renderQueryEditor()}</div>
    </div>
  );
};
