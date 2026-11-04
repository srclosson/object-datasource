import React, { useEffect, useState } from 'react';
import { Field, Input, SecretInput } from '@grafana/ui';
import { DataSourcePluginOptionsEditorProps } from '@grafana/data';
import { CacheSettings, ObjectDataSourceOptions, ObjectSecureJsonData } from '../types';
import { QueryLinks } from './QueryLinks';
import { savedNames } from '../tree';

interface Props extends DataSourcePluginOptionsEditorProps<ObjectDataSourceOptions, ObjectSecureJsonData> {}

export const ConfigEditor = ({ options, onOptionsChange }: Props) => {
  const { jsonData, secureJsonFields, secureJsonData } = options;

  // Where each link was at the last save. Grafana bumps the version on save, which moves new
  // links into their folders and closes everything again.
  const [saved, setSaved] = useState(() => savedNames(jsonData.queryLinks));
  useEffect(() => {
    setSaved(savedNames(options.jsonData.queryLinks));
    // Only on save; the links change on every edit
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [options.version]);

  const setCache = (patch: Partial<CacheSettings>) =>
    onOptionsChange({ ...options, jsonData: { ...jsonData, cache: { ...jsonData.cache, ...patch } } });

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
      <h3 className="page-heading">Query cache</h3>
      <Field
        label="Cache results for (seconds)"
        description="How long query results are reused, unless a folder or link says otherwise. Refreshes of a moving range like the last hour share results within this time. Alert rules always get fresh data. Negative turns the cache off."
      >
        <Input
          width={20}
          type="number"
          placeholder="30"
          value={jsonData.cache?.defaultTtl ?? ''}
          onChange={(e) => {
            const v = e.currentTarget.value;
            setCache({ defaultTtl: v === '' ? undefined : Number(v) });
          }}
        />
      </Field>
      <Field
        label="Redis address"
        description="host:port of a Redis server, so the cache survives restarts and is shared by every Grafana. Leave empty to cache in memory."
      >
        <Input
          width={50}
          placeholder="redis:6379"
          value={jsonData.cache?.redisAddress ?? ''}
          onChange={(e) => setCache({ redisAddress: e.currentTarget.value })}
        />
      </Field>
      <Field label="Redis password">
        <SecretInput
          width={50}
          isConfigured={Boolean(secureJsonFields?.redisPassword)}
          value={secureJsonData?.redisPassword ?? ''}
          onChange={(e) =>
            onOptionsChange({
              ...options,
              secureJsonData: { ...secureJsonData, redisPassword: e.currentTarget.value },
            })
          }
          onReset={() =>
            onOptionsChange({
              ...options,
              secureJsonFields: { ...secureJsonFields, redisPassword: false },
              secureJsonData: { ...secureJsonData, redisPassword: '' },
            })
          }
        />
      </Field>
      <QueryLinks
        key={options.version}
        saved={saved}
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
