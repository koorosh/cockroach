// Copyright 2019 The Cockroach Authors.
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

import * as React from "react";
import { Icon, Select } from "antd";

import "./metricsSelect.styl";

export interface MetricsSelectProps {
  value: any;
  onChange: (value: string) => void;
  options: MetricOption[];
}

export interface MetricOption {
  label: string;
  value: string;
  description: string;
}

const {Option} = Select;

export function MetricsSelect(props: MetricsSelectProps) {
  const {value, onChange, options} = props;
  return (
    <div className="metric-select">
      <Select
        className="metric-select__input"
        showSearch
        placeholder="Select a metric..."
        optionFilterProp="title"
        onChange={ onChange }
        optionLabelProp="value"
        filterOption={ true }
        allowClear
        showArrow={ false }
        size="large"
        value={ value }
        dropdownMatchSelectWidth={ false }
        dropdownClassName="metric-select__dropdown-menu"
        notFoundContent="No results found"
      >
        {
          options.map(option => (<Option value={ option.value } title={ option.label }>
            <div title={ option.description } className="metric-select__option">
              <div className="metric-select__option-label">{ option.label }</div>
              <div
                className="metric-select__description">
                { option.description }
              </div>
            </div>
          </Option>))
        }
      </Select>
      <Icon type="search" className="metric-select--search-icon"/>
    </div>
  );
}
