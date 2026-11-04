import React from 'react';
import { css } from '@emotion/css';
import { DataSourceInstanceSettings, GrafanaTheme2 } from '@grafana/data';
import { DataSourcePicker } from '@grafana/runtime';
import {
  Button,
  ColorPickerInput,
  ControlledCollapse,
  IconButton,
  InlineField,
  InlineFieldRow,
  Input,
  TextArea,
  UnitPicker,
  useStyles2,
} from '@grafana/ui';
import { LabelLookup, Metadata } from '../types';
import { DatasourceQueryEditor } from './DatasourceQueryEditor';
import { Preview } from './Preview';

type Props = {
  label: string;
  value?: Metadata;
  /** What the node inherits from the folders above it, shown as placeholders */
  inherited?: Metadata;
  onChange: (value: Metadata) => void;
};

const getStyles = (theme: GrafanaTheme2) => ({
  section: css`
    margin-top: ${theme.spacing(2)};
  `,
  heading: css`
    font-weight: ${theme.typography.fontWeightMedium};
    margin-bottom: ${theme.spacing(0.5)};
  `,
  hint: css`
    color: ${theme.colors.text.secondary};
    font-size: ${theme.typography.bodySmall.fontSize};
    margin-bottom: ${theme.spacing(1)};
  `,
  lookup: css`
    border-left: 2px solid ${theme.colors.border.medium};
    padding-left: ${theme.spacing(1)};
    margin-bottom: ${theme.spacing(2)};
  `,
});

const LABEL_WIDTH = 16;

/** Edits the metadata of a folder or query link */
export const MetadataEditor = ({ label, value = {}, inherited = {}, onChange }: Props) => {
  const styles = useStyles2(getStyles);
  const set = (patch: Partial<Metadata>) => onChange({ ...value, ...patch });
  const labels = Object.entries(value.labels ?? {});
  const inheritedNote = (v?: string | number) => (v === undefined || v === '' ? undefined : `Inherited: ${v}`);

  const setLabels = (entries: Array<[string, string]>) => set({ labels: Object.fromEntries(entries) });
  const setLookup = (i: number, lookup: LabelLookup) => {
    const lookups = [...(value.labelLookups ?? [])];
    lookups[i] = lookup;
    set({ labelLookups: lookups });
  };

  return (
    <ControlledCollapse label={label} isOpen={false}>
      <InlineField label="Description" labelWidth={LABEL_WIDTH} grow tooltip="What the data means">
        <TextArea
          rows={2}
          value={value.description ?? ''}
          placeholder={inheritedNote(inherited.description)}
          onChange={(e) => set({ description: e.currentTarget.value })}
        />
      </InlineField>
      <InlineFieldRow>
        <InlineField label="Unit" labelWidth={LABEL_WIDTH}>
          <UnitPicker value={value.unit} onChange={(unit) => set({ unit })} />
        </InlineField>
        {inherited.unit && !value.unit && <div className={styles.hint}>Inherited: {inherited.unit}</div>}
      </InlineFieldRow>
      <InlineField
        label="Display name"
        labelWidth={LABEL_WIDTH}
        tooltip="What series are called in panels. Can use labels, e.g. Temperature ${__field.labels.host}"
      >
        <Input
          width={50}
          value={value.displayName ?? ''}
          placeholder={inheritedNote(inherited.displayName)}
          onChange={(e) => set({ displayName: e.currentTarget.value })}
        />
      </InlineField>
      <InlineFieldRow>
        <InlineField label="Min" labelWidth={LABEL_WIDTH} tooltip="The lowest expected value">
          <NumberInput value={value.min} placeholder={inheritedNote(inherited.min)} onChange={(min) => set({ min })} />
        </InlineField>
        <InlineField label="Max" tooltip="The highest expected value">
          <NumberInput value={value.max} placeholder={inheritedNote(inherited.max)} onChange={(max) => set({ max })} />
        </InlineField>
      </InlineFieldRow>

      <div className={styles.section}>
        <div className={styles.heading}>Thresholds</div>
        <div className={styles.hint}>Levels at which the data needs attention; panels color themselves by them.</div>
        {(value.thresholds ?? []).map((t, i) => (
          <InlineFieldRow key={i}>
            <InlineField label="At or above" labelWidth={LABEL_WIDTH}>
              <NumberInput
                value={t.value}
                onChange={(v) => set({ thresholds: replaceAt(value.thresholds!, i, { ...t, value: v ?? 0 }) })}
              />
            </InlineField>
            <InlineField label="Color">
              <ColorPickerInput
                value={t.color}
                onChange={(color) => set({ thresholds: replaceAt(value.thresholds!, i, { ...t, color }) })}
              />
            </InlineField>
            <IconButton name="trash-alt" tooltip="Remove threshold" onClick={() => set({ thresholds: removeAt(value.thresholds!, i) })} />
          </InlineFieldRow>
        ))}
        <Button
          size="sm"
          variant="secondary"
          icon="plus"
          onClick={() => {
            const last = value.thresholds?.at(-1)?.value;
            set({ thresholds: [...(value.thresholds ?? []), { value: last === undefined ? 80 : last + 10, color: 'red' }] });
          }}
        >
          Add threshold
        </Button>
      </div>

      <div className={styles.section}>
        <div className={styles.heading}>Labels</div>
        <div className={styles.hint}>
          Added to every series, for grouping and alert routing.
          {inherited.labels && Object.keys(inherited.labels).length > 0 &&
            ` Inherited: ${Object.entries(inherited.labels)
              .map(([k, v]) => `${k}=${v}`)
              .join(', ')}`}
        </div>
        {labels.map(([k, v], i) => (
          <InlineFieldRow key={i}>
            <InlineField label="Name" labelWidth={LABEL_WIDTH}>
              <Input width={24} value={k} onChange={(e) => setLabels(replaceAt(labels, i, [e.currentTarget.value, v]))} />
            </InlineField>
            <InlineField label="Value">
              <Input width={30} value={v} onChange={(e) => setLabels(replaceAt(labels, i, [k, e.currentTarget.value]))} />
            </InlineField>
            <IconButton name="trash-alt" tooltip="Remove label" onClick={() => setLabels(removeAt(labels, i))} />
          </InlineFieldRow>
        ))}
        <Button size="sm" variant="secondary" icon="plus" onClick={() => setLabels([...labels, [uniqueName(labels), '']])}>
          Add label
        </Button>
      </div>

      <div className={styles.section}>
        <div className={styles.heading}>Labels from other datasources</div>
        <div className={styles.hint}>
          A query whose results become labels: a table&apos;s columns, or a series&apos; labels. Match on a label to give each
          series its own row (e.g. match host to an asset table); otherwise every series gets the first row.
        </div>
        {(value.labelLookups ?? []).map((lookup, i) => (
          <div key={i} className={styles.lookup}>
            <InlineFieldRow>
              <InlineField label="Datasource" labelWidth={LABEL_WIDTH}>
                <DataSourcePicker
                  current={lookup.uid}
                  onChange={(ds: DataSourceInstanceSettings) =>
                    setLookup(i, {
                      ...lookup,
                      uid: ds.uid,
                      query: { ...lookup.query, datasource: { uid: ds.uid, type: ds.type } },
                    })
                  }
                />
              </InlineField>
              <IconButton
                name="trash-alt"
                tooltip="Remove lookup"
                onClick={() => set({ labelLookups: removeAt(value.labelLookups!, i) })}
              />
            </InlineFieldRow>
            <DatasourceQueryEditor uid={lookup.uid} query={lookup.query} onChange={(query) => setLookup(i, { ...lookup, query })} />
            <InlineFieldRow>
              <InlineField label="Match label" labelWidth={LABEL_WIDTH} tooltip="The series label to join on, e.g. host">
                <Input
                  width={24}
                  value={lookup.matchLabel ?? ''}
                  placeholder="none: use the first row"
                  onChange={(e) => setLookup(i, { ...lookup, matchLabel: e.currentTarget.value })}
                />
              </InlineField>
              <InlineField label="to column" tooltip="The lookup's column holding that label's values. Defaults to the label's name.">
                <Input
                  width={24}
                  value={lookup.matchColumn ?? ''}
                  placeholder={lookup.matchLabel || 'same name'}
                  onChange={(e) => setLookup(i, { ...lookup, matchColumn: e.currentTarget.value })}
                />
              </InlineField>
            </InlineFieldRow>
            <Preview uid={lookup.uid} query={lookup.query} mode="labels" />
          </div>
        ))}
        <Button
          size="sm"
          variant="secondary"
          icon="plus"
          onClick={() => set({ labelLookups: [...(value.labelLookups ?? []), { query: { refId: 'lookup' } }] })}
        >
          Add label lookup
        </Button>
      </div>

      <div className={styles.section}>
        <div className={styles.heading}>Links</div>
        <div className={styles.hint}>
          Runbooks, explanations, manufacturers&apos; pages: clickable on the data in every panel. Links accumulate down the tree.
        </div>
        {(value.links ?? []).map((link, i) => (
          <InlineFieldRow key={i}>
            <InlineField label="Title" labelWidth={LABEL_WIDTH}>
              <Input
                width={24}
                value={link.title}
                onChange={(e) => set({ links: replaceAt(value.links!, i, { ...link, title: e.currentTarget.value }) })}
              />
            </InlineField>
            <InlineField label="URL" grow>
              <Input
                value={link.url}
                placeholder="https://"
                onChange={(e) => set({ links: replaceAt(value.links!, i, { ...link, url: e.currentTarget.value }) })}
              />
            </InlineField>
            <IconButton
              name="external-link-alt"
              tooltip="Open"
              disabled={!/^https?:\/\//.test(link.url)}
              onClick={() => window.open(link.url, '_blank', 'noopener,noreferrer')}
            />
            <IconButton name="trash-alt" tooltip="Remove link" onClick={() => set({ links: removeAt(value.links!, i) })} />
          </InlineFieldRow>
        ))}
        <Button
          size="sm"
          variant="secondary"
          icon="plus"
          onClick={() => set({ links: [...(value.links ?? []), { title: '', url: '' }] })}
        >
          Add link
        </Button>
      </div>
    </ControlledCollapse>
  );
};

const NumberInput = ({
  value,
  placeholder,
  onChange,
}: {
  value?: number;
  placeholder?: string;
  onChange: (v: number | undefined) => void;
}) => (
  <Input
    width={16}
    type="number"
    value={value ?? ''}
    placeholder={placeholder}
    onChange={(e) => {
      const v = e.currentTarget.value;
      onChange(v === '' ? undefined : Number(v));
    }}
  />
);

const replaceAt = <T,>(list: T[], i: number, item: T): T[] => list.map((x, j) => (j === i ? item : x));
const removeAt = <T,>(list: T[], i: number): T[] => list.filter((_, j) => j !== i);

/** A label name not used yet, so adding a label never merges into an existing one */
const uniqueName = (labels: Array<[string, string]>): string => {
  const names = new Set(labels.map(([k]) => k));
  let i = labels.length + 1;
  while (names.has(`label${i}`)) {
    i++;
  }
  return `label${i}`;
};
