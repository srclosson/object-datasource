import React, { useState } from 'react';
import { css } from 'emotion';
import {
  VariableSuggestion,
  DataSourceInstanceSettings,
  DataSourceJsonData,
  DataSourceApi,
  PanelData,
  DataQuery,
  LoadingState,
} from '@grafana/data';
import { Button, LegacyForms, stylesFactory } from '@grafana/ui';
import { getLegacyAngularInjector, getDataSourceSrv } from '@grafana/runtime';
const { FormField } = LegacyForms;
import { QueryLinkConfig } from '../types';
import { DataSourcePicker } from 'ui-enterprise';

const getStyles = stylesFactory(() => ({
  firstRow: css`
    display: flex;
  `,
  nameField: css`
    flex: 2;
  `,
  regexField: css`
    flex: 3;
  `,
  row: css`
    display: flex;
    align-items: baseline;
  `,
  editor: css`
    display: flex;
    flex-direction: column;
    height: 100%;
  `,
}));

type Props = {
  value: QueryLinkConfig;
  datasources?: Array<DataSourceInstanceSettings<DataSourceJsonData>>;
  onChange: (value: QueryLinkConfig) => void;
  onDelete: () => void;
  suggestions: VariableSuggestion[];
  className?: string;
};

type State = {
  queryString: string;
  datasource: DataSourceApi | null;
  query: DataQuery;
};

export const QueryLink = (props: Props) => {
  const { value, onChange, onDelete, className } = props;
  const [state, setState] = useState<State>({
    queryString: value.query ? JSON.stringify(value.query) : '',
    datasource: null,
    query: {
      refId: '',
    },
  });

  const styles = getStyles();

  const handleChange = (field: keyof typeof value) => (event: React.ChangeEvent<HTMLInputElement>) => {
    onChange({
      ...value,
      [field]: event.currentTarget.value,
    });
  };

  const onQueryChange = (query: any) => {
    console.log('We just got a query', query);
    onChange({
      ...props.value,
      query,
    });
  };

  const onRunQuery = () => {
    console.log('We just got an onRunQuery event');
  };

  const renderPluginEditor = () => {
    const { value } = props;
    const { query } = value;
    const { datasource } = state;
    const timeRange = getLegacyAngularInjector().get('timeSrv').timeRange();
    const data: PanelData = {
      state: LoadingState.NotStarted,
      series: [],
      timeRange,
    };

    if (datasource?.components?.QueryEditor && query) {
      const QueryEditor = datasource.components.QueryEditor;

      return (
        <QueryEditor
          key={datasource?.name}
          query={query}
          datasource={datasource}
          onChange={onQueryChange}
          onRunQuery={onRunQuery}
          data={data}
          range={timeRange}
          queries={[query]}
        />
      );
    }

    return <div>Data source plugin does not export any Query Editor component</div>;
  };

  const loadDatasource = async (config?: QueryLinkConfig) => {
    const { onChange } = props;
    const dataSourceSrv = getDataSourceSrv();
    let datasource: DataSourceApi;

    if (config?.uid) {
      datasource = await dataSourceSrv.get(config.uid);
    } else {
      // Got the default datasource. Write out an onchange event
      datasource = await dataSourceSrv.get();
      onChange({
        name: '',
        uid: datasource.uid,
        query: {
          refId: '',
          datasource: datasource.name,
          datasourceId: datasource.id,
        },
      });
    }

    setState({
      ...state,
      datasource,
    });
  };

  if (!state.datasource) {
    loadDatasource(props.value);
  }

  return (
    <div className={className}>
      <div className={styles.firstRow + ' gf-form'}>
        <FormField
          className={styles.nameField}
          labelWidth={6}
          // A bit of a hack to prevent using default value for the width from FormField
          inputWidth={null}
          label="Name"
          type="text"
          value={value.name}
          tooltip={
            'Think of this a a folder name. Columns in the query returned will be underneath this name as the folder name'
          }
          onChange={handleChange('name')}
        />
        <Button
          variant={'destructive'}
          title="Remove field"
          icon="times"
          onClick={(event) => {
            event.preventDefault();
            onDelete();
          }}
        />
      </div>
      <div className="gf-form">
        <FormField
          label={'Datasource'}
          labelWidth={6}
          inputEl={
            <DataSourcePicker
              // Uid and value should be always set in the db and so in the items.
              onChange={(ds) => {
                console.log('new ds', ds);
                onChange({
                  ...value,
                  uid: ds.uid,
                  query: {
                    ...value.query,
                    refId: value.name || '',
                    datasourceId: ds.id,
                    datasource: ds.name,
                  },
                });
                loadDatasource(ds);
              }}
              current={value.uid}
            />
          }
          className={css`
            width: 100%;
          `}
        />
      </div>
      <div className={styles.editor}>
        <div className="query-editor-row">
          <div className="gf-form-group">{renderPluginEditor()}</div>
        </div>
      </div>
      <div className={styles.row}></div>
    </div>
  );
};
