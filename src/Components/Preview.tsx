import React, { useEffect, useMemo, useRef, useState } from 'react';
import { css } from '@emotion/css';
import { lastValueFrom } from 'rxjs';
import {
  DataFrame,
  DataQueryRequest,
  DataQueryResponse,
  FieldType,
  GrafanaTheme2,
  applyFieldOverrides,
  getDefaultTimeRange,
  getFrameDisplayName,
  toDataFrame,
} from '@grafana/data';
import { DataQuery } from '@grafana/schema';
import { getDataSourceSrv } from '@grafana/runtime';
import { Alert, Button, Combobox, Table, useStyles2, useTheme2 } from '@grafana/ui';
import { labelRows } from '../metadata';
import { variableNames } from '../variables';

type Props = {
  uid?: string;
  query: DataQuery;
  /** "data" shows the results; "labels" shows the labels a label lookup reads from them */
  mode: 'data' | 'labels';
};

const getStyles = (theme: GrafanaTheme2) => ({
  container: css`
    margin: ${theme.spacing(1, 0)};
  `,
  summary: css`
    display: flex;
    align-items: center;
    gap: ${theme.spacing(1)};
    margin: ${theme.spacing(1, 0)};
    color: ${theme.colors.text.secondary};
  `,
});

/**
 * Runs a query against its datasource over the last 6 hours and shows the results in a table.
 * It runs in the browser as the person editing, not as the plugin's service account.
 */
export const Preview = ({ uid, query, mode }: Props) => {
  const styles = useStyles2(getStyles);
  const theme = useTheme2();
  const [state, setState] = useState<{ loading?: boolean; frames?: DataFrame[]; error?: string }>({});
  const [selected, setSelected] = useState(0);
  const ref = useRef<HTMLDivElement>(null);
  const width = useWidth(ref);

  const run = async () => {
    if (!uid) {
      return;
    }
    setState({ loading: true });
    setSelected(0);
    try {
      const ds = await getDataSourceSrv().get(uid);
      const request = {
        requestId: `object-datasource-preview-${Date.now()}`,
        app: 'object-datasource-preview',
        interval: '1m',
        intervalMs: 60000,
        maxDataPoints: 500,
        range: getDefaultTimeRange(),
        scopedVars: {},
        targets: [{ ...query, refId: 'Preview' }],
        timezone: 'browser',
        startTime: Date.now(),
      } as DataQueryRequest<DataQuery>;
      const result = ds.query(request);
      const res: DataQueryResponse | undefined = 'then' in result ? await result : await lastValueFrom(result);
      const error = res?.errors?.[0]?.message;
      setState({ frames: (res?.data ?? []).map((d) => toDataFrame(d)), error });
    } catch (e) {
      setState({ error: e instanceof Error ? e.message : String(e) });
    }
  };

  const table = useMemo(() => {
    const frames = state.frames ?? [];
    let frame: DataFrame | undefined;
    if (mode === 'labels') {
      const rows = labelRows(frames);
      const keys = [...new Set(rows.flatMap((r) => Object.keys(r)))];
      frame = rows.length
        ? toDataFrame({
            fields: keys.map((k) => ({ name: k, type: FieldType.string, values: rows.map((r) => r[k] ?? '') })),
          })
        : undefined;
    } else {
      frame = frames[selected];
    }
    if (!frame || frame.length === 0) {
      return undefined;
    }
    return applyFieldOverrides({
      data: [frame],
      fieldConfig: { defaults: {}, overrides: [] },
      replaceVariables: (v: string) => v,
      theme,
      timeZone: 'browser',
    })[0];
  }, [state.frames, selected, mode, theme]);

  const frames = state.frames;
  const variables = variableNames(query);
  return (
    <div className={styles.container} ref={ref}>
      <div className={styles.summary}>
        <Button size="sm" variant="secondary" icon={state.loading ? 'spinner' : 'table'} disabled={!uid} onClick={run}>
          {frames ? 'Run again' : 'Preview'}
        </Button>
        {frames && mode === 'data' && <span>{frames.length} frame(s), last 6 hours</span>}
        {frames && mode === 'labels' && <span>{labelRows(frames).length} row(s) of labels</span>}
        {frames && mode === 'data' && frames.length > 1 && (
          <Combobox
            width={50}
            value={selected}
            options={frames.map((f, i) => ({ value: i, label: getFrameDisplayName(f, i) }))}
            onChange={(v) => setSelected(v.value)}
          />
        )}
        {frames && (
          <Button size="sm" variant="secondary" fill="text" icon="times" onClick={() => setState({})}>
            Close
          </Button>
        )}
      </div>
      {state.error && (
        <Alert severity="error" title="The query failed">
          {state.error}
        </Alert>
      )}
      {frames && !state.error && !table && (
        <div className={styles.summary}>
          No data
          {variables.length > 0 &&
            `. The query uses dashboard variables (${variables.map((v) => '$' + v).join(', ')}), which have no value here.`}
        </div>
      )}
      {table && width > 0 && <Table data={table} width={width} height={240} />}
    </div>
  );
};

const useWidth = (ref: React.RefObject<HTMLElement>) => {
  const [width, setWidth] = useState(0);
  useEffect(() => {
    if (!ref.current) {
      return;
    }
    const observer = new ResizeObserver((entries) => setWidth(entries[0].contentRect.width));
    observer.observe(ref.current);
    return () => observer.disconnect();
  }, [ref]);
  return width;
};
