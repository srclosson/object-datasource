import React, { Component } from 'react';
import { PanelProps } from '@grafana/data';
import { ObjectPanelType } from 'types';

interface Props extends PanelProps<ObjectPanelType> {}

export class ObjectPanel extends Component<Props> {
  constructor(props: Props) {
    super(props);
  }

  render() {
    return <></>;
  }
}
