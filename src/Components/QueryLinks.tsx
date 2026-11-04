import React from 'react';
import { css } from '@emotion/css';
import { Button, ControlledCollapse, useStyles2 } from '@grafana/ui';
import { GrafanaTheme2 } from '@grafana/data';
import { getDataSourceSrv } from '@grafana/runtime';
import { Metadata, QueryLink, newLinkId } from '../types';
import { inherit } from '../metadata';
import { FolderNode, groupLinks } from '../tree';
import { MetadataEditor } from './MetadataEditor';
import { QueryLinkEditor } from './QueryLink';

const getStyles = (theme: GrafanaTheme2) => ({
  infoText: css`
    padding-bottom: ${theme.spacing(2)};
    color: ${theme.colors.text.secondary};
  `,
  newLink: css`
    margin-bottom: ${theme.spacing(2)};
    padding-bottom: ${theme.spacing(2)};
    border-bottom: 1px solid ${theme.colors.border.weak};
  `,
  folder: css`
    padding-left: ${theme.spacing(2)};
  `,
  count: css`
    margin-left: ${theme.spacing(1)};
    color: ${theme.colors.text.secondary};
    font-weight: ${theme.typography.fontWeightRegular};
  `,
});

type Props = {
  value?: QueryLink[];
  folders?: Record<string, Metadata>;
  /** The name each link had at the last save, which places it in the tree */
  saved: Map<string, string>;
  onChange: (value: QueryLink[]) => void;
  onFoldersChange: (folders: Record<string, Metadata>) => void;
};

/**
 * The query links, arranged as the tree people pick from. Links added since the last save are
 * listed first, open for editing; on save they move into their folders, which start closed.
 */
export const QueryLinks = ({ value = [], folders = {}, saved, onChange, onFoldersChange }: Props) => {
  const styles = useStyles2(getStyles);
  const { fresh, root } = groupLinks(value, saved);

  const update = (index: number, link: QueryLink) => {
    const links = [...value];
    // Links saved before they had ids get one the first time they're edited
    links.splice(index, 1, { ...link, id: link.id ?? newLinkId() });
    onChange(links);
  };
  const remove = (index: number) => onChange(value.filter((_, i) => i !== index));

  const renderLink = (index: number) => (
    <QueryLinkEditor
      value={value[index]}
      folders={folders}
      onChange={(link) => update(index, link)}
      onDelete={() => remove(index)}
    />
  );

  const renderFolder = (folder: FolderNode): React.ReactNode => (
    <>
      {folder.path && (
        <MetadataEditor
          label="Folder metadata"
          value={folders[folder.path]}
          inherited={inheritedBy(folder.path, folders)}
          onChange={(md) => onFoldersChange({ ...folders, [folder.path]: md })}
        />
      )}
      {folder.folders.map((child) => (
        <ControlledCollapse
          key={child.path}
          isOpen={false}
          label={
            <>
              {child.label}
              <span className={styles.count}>{countLinks(child)} link(s)</span>
            </>
          }
        >
          <div className={styles.folder}>{renderFolder(child)}</div>
        </ControlledCollapse>
      ))}
      {folder.links.map((index) => (
        <ControlledCollapse key={value[index].id ?? index} isOpen={false} label={linkLabel(value[index], styles.count)}>
          {renderLink(index)}
        </ControlledCollapse>
      ))}
    </>
  );

  return (
    <>
      <h3 className="page-heading">Query links</h3>
      <div className={styles.infoText}>
        Add queries for existing datasources. Dots in a name place it in the tree, like folders; on save, new links
        move into their folders.
      </div>

      {fresh.map((index) => (
        <div key={value[index].id ?? index} className={styles.newLink}>
          {renderLink(index)}
        </div>
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

      <div className="gf-form-group">{renderFolder(root)}</div>
    </>
  );
};

const countLinks = (folder: FolderNode): number =>
  folder.links.length + folder.folders.reduce((n, f) => n + countLinks(f), 0);

const linkLabel = (link: QueryLink, countClass: string) => {
  const ds = link.uid ? getDataSourceSrv().getInstanceSettings(link.uid)?.name : undefined;
  const name = link.name?.split('.').pop() || '(unnamed)';
  return (
    <>
      {name}
      {ds && <span className={countClass}>{ds}</span>}
    </>
  );
};

/** What a folder inherits from the folders above it */
const inheritedBy = (path: string, folders: Record<string, Metadata>): Metadata => {
  const segments = path.split('.');
  let md: Metadata = {};
  for (let i = 1; i < segments.length; i++) {
    md = inherit(folders[segments.slice(0, i).join('.')], md);
  }
  return md;
};
