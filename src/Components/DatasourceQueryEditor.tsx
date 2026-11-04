import React, { useEffect, useMemo, useState } from 'react';
import { DataSourceApi, LoadingState, PanelData, getDefaultTimeRange } from '@grafana/data';
import { DataQuery } from '@grafana/schema';
import { getDataSourceSrv } from '@grafana/runtime';

type Props = {
  uid?: string;
  query: DataQuery;
  onChange: (query: DataQuery) => void;
};

/** A datasource's own query editor, for writing a query against it */
export const DatasourceQueryEditor = ({ uid, query, onChange }: Props) => {
  const [datasource, setDatasource] = useState<DataSourceApi>();

  useEffect(() => {
    let cancelled = false;
    if (uid) {
      getDataSourceSrv()
        .get(uid)
        .then((ds) => !cancelled && setDatasource(ds))
        .catch(() => !cancelled && setDatasource(undefined));
    }
    return () => {
      cancelled = true;
    };
  }, [uid]);

  const data: PanelData = useMemo(
    () => ({ state: LoadingState.NotStarted, series: [], timeRange: getDefaultTimeRange() }),
    []
  );

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
      query={query}
      datasource={datasource}
      onChange={onChange}
      onRunQuery={() => {}}
      data={data}
      range={data.timeRange}
      queries={[query]}
    />
  );
};
