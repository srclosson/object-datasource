import React from 'react';
import { css } from '@emotion/css';
import { Button, useStyles2 } from '@grafana/ui';
import { GrafanaTheme2 } from '@grafana/data';
import { QueryLink, newLinkId } from '../types';
import { QueryLinkEditor } from './QueryLink';

const getStyles = (theme: GrafanaTheme2) => ({
  infoText: css`
    padding-bottom: ${theme.spacing(2)};
    color: ${theme.colors.text.secondary};
  `,
  dataLink: css`
    margin-bottom: ${theme.spacing(2)};
  `,
});

type Props = {
  value?: QueryLink[];
  onChange: (value: QueryLink[]) => void;
};

export const QueryLinks = ({ value = [], onChange }: Props) => {
  const styles = useStyles2(getStyles);

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
    </>
  );
};
