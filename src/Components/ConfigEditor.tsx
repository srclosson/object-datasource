import React from 'react';
import { Field, Input, SecretInput } from '@grafana/ui';
import { DataSourcePluginOptionsEditorProps } from '@grafana/data';
import { ObjectDataSourceOptions, ObjectSecureJsonData } from '../types';
import { QueryLinks } from './QueryLinks';

interface Props extends DataSourcePluginOptionsEditorProps<ObjectDataSourceOptions, ObjectSecureJsonData> {}

export const ConfigEditor = ({ options, onOptionsChange }: Props) => {
  const { jsonData, secureJsonFields, secureJsonData } = options;

  return (
    <>
      <h3 className="page-heading">Grafana API</h3>
      <Field
        label="Grafana URL"
        description="Where the plugin reaches Grafana's API. Leave empty to use Grafana's root_url."
      >
        <Input
          width={50}
          placeholder="http://localhost:3000"
          value={options.url}
          onChange={(e) => onOptionsChange({ ...options, url: e.currentTarget.value })}
        />
      </Field>
      <Field
        label="Service account token"
        description="Used to run query links against other datasources. Leave empty when Grafana gives the plugin a managed service account (the externalServiceAccounts feature)."
      >
        <SecretInput
          width={50}
          placeholder="glsa_..."
          isConfigured={Boolean(secureJsonFields?.serviceAccountToken)}
          value={secureJsonData?.serviceAccountToken ?? ''}
          onChange={(e) =>
            onOptionsChange({
              ...options,
              secureJsonData: { ...secureJsonData, serviceAccountToken: e.currentTarget.value },
            })
          }
          onReset={() =>
            onOptionsChange({
              ...options,
              secureJsonFields: { ...secureJsonFields, serviceAccountToken: false },
              secureJsonData: { ...secureJsonData, serviceAccountToken: '' },
            })
          }
        />
      </Field>
      <QueryLinks
        value={jsonData.queryLinks || []}
        folders={jsonData.folders}
        onChange={(queryLinks) =>
          onOptionsChange({
            ...options,
            jsonData: { ...jsonData, queryLinks },
          })
        }
        onFoldersChange={(folders) =>
          onOptionsChange({
            ...options,
            jsonData: { ...jsonData, folders },
          })
        }
      />
    </>
  );
};
