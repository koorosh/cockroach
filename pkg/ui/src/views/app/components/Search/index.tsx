// Copyright 2018 The Cockroach Authors.
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

import React from "react";
import { Button, Form, Input, ConfigProvider } from "antd";
import { InputProps } from "antd/lib/input";
import { ANT_DESIGN_PREFIX } from "src/util/constants";
import CancelIcon from "assets/cancel.svg";
import SearchIcon from "assets/search.svg";
import "antd/es/input/style";
import "antd/es/form/style";
import "antd/es/button/style";
import styles from  "./search.module.styl";

interface ISearchProps {
  onSubmit: (value: string) => void;
  onClear?: () => void;
  defaultValue?: string;
}

interface ISearchState {
  value: string;
  submitted: boolean;
  submit?: boolean;
}

type TSearchProps = ISearchProps & InputProps;

export class Search extends React.Component<TSearchProps, ISearchState> {
  state = {
    value: this.props.defaultValue || "",
    submitted: false,
  };

  onSubmit = () => {
    const { value } = this.state;
    const { onSubmit } = this.props;
    onSubmit(value);
    if (value.length > 0) {
      this.setState({
        submitted: true,
      });
    }
  }

  onChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    const value: string = event.target.value;
    const submitted = value.length === 0;
    this.setState({ value, submitted }, () => submitted && this.onSubmit());
  }

  onClear = () => {
    const { onClear } = this.props;
    this.setState({ value: "", submit: false });
    onClear();
  }

  renderSuffix = () => {
    const { value, submitted } = this.state;
    if (value.length > 0) {
      if (submitted) {
        return (
          <Button
            onClick={this.onClear}
            type="default"
            className={styles._clearSearch}
          >
            <img className={styles._suffixIcon} src={CancelIcon} />
          </Button>
        );
      }
      return (
        <Button
          type="default"
          htmlType="submit"
          className={styles._submitSearch}
        >
          Enter
        </Button>
      );
    }
    return null;
  }

  render() {
    const { value, submitted } = this.state;
    const className = submitted ? styles._submitted : "";
    return (
      <ConfigProvider prefixCls={ANT_DESIGN_PREFIX}>
        <Form onSubmit={this.onSubmit} className={styles._searchForm}>
          <Form.Item>
            <Input
              className={className}
              placeholder="Search Statement"
              onChange={this.onChange}
              prefix={<img className={styles._prefixIcon} src={SearchIcon} />}
              suffix={this.renderSuffix()}
              value={value}
              {...this.props}
           />
          </Form.Item>
        </Form>
      </ConfigProvider>
    );
  }
}
