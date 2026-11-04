import React from 'react';
import { css } from '@emotion/css';
import { Button, useStyles2 } from '@grafana/ui';
import { GrafanaTheme2 } from '@grafana/data';
import { Metadata, QueryLink, newLinkId } from '../types';
import { folderPaths, inherit } from '../metadata';
import { MetadataEditor } from './MetadataEditor';
import { QueryLinkEditor } from './QueryLink';

const getStyles = (theme: GrafanaTheme2) => ({
  infoText: css`
    padding-bottom: ${theme.spacing(2)};
    color: ${theme.colors.text.secondary};
  `,
  dataLink: css`
    margin-bottom: ${theme.spacing(2)};
    padding-bottom: ${theme.spacing(2)};
    border-bottom: 1px solid ${theme.colors.border.weak};
  `,
});

type Props = {
  value?: QueryLink[];
  folders?: Record<string, Metadata>;
  onChange: (value: QueryLink[]) => void;
  onFoldersChange: (folders: Record<string, Metadata>) => void;
};

export const QueryLinks = ({ value = [], folders = {}, onChange, onFoldersChange }: Props) => {
  const styles = useStyles2(getStyles);
  const paths = folderPaths(value);

  return (
    <>
      <h3 className="page-heading">Query links</h3>

      <div className={styles.infoText}>
        Add queries for existing datasources. Dots in a name place it in the tree, like folders.
      </div>

      <div className="gf-form-group">
        {value.map((link, index) => (
          <QueryLinkEditor
            className={styles.dataLink}
            key={link.id ?? index}
            value={link}
            folders={folders}
            onChange={(newLink) => {
              const links = [...value];
              // Links saved before they had ids get one the first time they're edited
              links.splice(index, 1, { ...newLink, id: newLink.id ?? newLinkId() });
              onChange(links);
            }}
            onDelete={() => {
              const links = [...value];
              links.splice(index, 1);
              onChange(links);
            }}
          />
        ))}
        <Button
          variant="secondary"
          icon="plus"
          onClick={(event) => {
            event.preventDefault();
            onChange([...value, { id: newLinkId(), query: { refId: '' } }]);
          }}
        >
          Add
        </Button>
      </div>

      {paths.length > 0 && (
        <>
          <h3 className="page-heading">Folders</h3>
          <div className={styles.infoText}>
            Metadata set on a folder applies to every link under it. A link&apos;s own values win; labels merge and links
            accumulate.
          </div>
          <div className="gf-form-group">
            {paths.map((path) => {
              // What the folder inherits from the folders above it
              const segments = path.split('.');
              let inherited: Metadata = {};
              for (let i = 1; i < segments.length; i++) {
                inherited = inherit(folders[segments.slice(0, i).join('.')], inherited);
              }
              return (
                <MetadataEditor
                  key={path}
                  label={path}
                  value={folders[path]}
                  inherited={inherited}
                  onChange={(md) => onFoldersChange({ ...folders, [path]: md })}
                />
              );
            })}
          </div>
        </>
      )}
    </>
  );
};
