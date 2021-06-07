import React, { useState } from 'react';
import { css } from 'emotion';
import { Button, stylesFactory, useTheme } from '@grafana/ui';
import { GrafanaTheme, VariableOrigin, DataLinkBuiltInVars, DataSourceInstanceSettings } from '@grafana/data';
import { QueryLink } from '../types';
import { QueryLinkEditor } from './QueryLink';

const getStyles = stylesFactory((theme: GrafanaTheme) => ({
  infoText: css`
    padding-bottom: ${theme.spacing.md};
    color: ${theme.colors.textWeak};
  `,
  dataLink: css`
    margin-bottom: ${theme.spacing.sm};
  `,
}));

type Props = {
  value?: QueryLink[];
  onChange: (value: QueryLink[]) => void;
};
export const QueryLinks = (props: Props) => {
  const { value, onChange } = props;
  const [datasources, setDatasources] = useState<DataSourceInstanceSettings[]>([]);
  const theme = useTheme();
  const styles = getStyles(theme);

  if (!datasources?.length) {
    fetch('/api/datasources').then(async (resp: Response) => {
      const restDS = (await resp.json()) as any[];
      const newDS: DataSourceInstanceSettings[] = restDS.map((ds) => {
        return {
          name: ds.name,
          value: ds.type,
          meta: {
            id: ds.id,
            info: {
              logos: {
                small: ds.typeLogoUrl,
              },
            },
          },
          sort: '',
        } as unknown as DataSourceInstanceSettings;
      });

      setDatasources(newDS);
    });
  }

  return (
    <>
      <h3 className="page-heading">Query links</h3>

      <div className={styles.infoText}>Add queries for exisitng datasources</div>

      <div className="gf-form-group">
        {value &&
          value.map((field, index) => {
            return (
              <QueryLinkEditor
                className={styles.dataLink}
                key={index}
                value={field}
                onChange={(newField) => {
                  console.log('we got an onChange in QueryLinks.tsx', newField);
                  const newDataLinks = [...value];
                  newDataLinks.splice(index, 1, newField);
                  onChange(newDataLinks);
                }}
                onDelete={() => {
                  const newDataLinks = [...value];
                  newDataLinks.splice(index, 1);
                  onChange(newDataLinks);
                }}
                suggestions={[
                  {
                    value: DataLinkBuiltInVars.valueRaw,
                    label: 'Raw value',
                    documentation: 'Raw value of the field',
                    origin: VariableOrigin.Value,
                  },
                ]}
              />
            );
          })}
        <div>
          <Button
            variant={'secondary'}
            className={css`
              margin-right: 10px;
            `}
            icon="plus"
            onClick={async (event) => {
              event.preventDefault();
              onChange([
                ...(value || []),
                {
                  query: {
                    refId: '',
                  },
                },
              ]);
            }}
          >
            Add
          </Button>
        </div>
      </div>
    </>
  );
};
