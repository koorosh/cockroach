// Copyright 2023 The Cockroach Authors.
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

import React, { useContext } from "react";
import { Tooltip } from "antd";
import "antd/lib/tooltip/style";
import { CircleFilled } from "../icon";
import { DatabasesPageDataDatabase } from "./databasesPage";
import classNames from "classnames/bind";
import styles from "./databasesPage.module.scss";
import { EncodeDatabaseUri } from "../util";
import { Link } from "react-router-dom";
import { StackIcon } from "../icon/stackIcon";
import { CockroachCloudContext } from "../contexts";
import { checkInfoAvailable, getNetworkErrorMessage } from "../databases";
import * as format from "../util/format";
import { Caution } from "@cockroachlabs/icons";

const cx = classNames.bind(styles);

interface CellProps {
  database: DatabasesPageDataDatabase;
}

export const DiskSizeCell = ({ database }: CellProps): JSX.Element => {
  return (
    <>
      {checkInfoAvailable(
        database.error,
        null,
        database.details?.stats?.approximate_disk_bytes
          ? format.Bytes(
              database.details?.stats?.approximate_disk_bytes?.toNumber(),
            )
          : null,
      )}
    </>
  );
};

export const IndexRecCell = ({ database }: CellProps): JSX.Element => {
  const text =
    database.details?.stats?.num_index_recommendations > 0
      ? `${database.details?.stats?.num_index_recommendations} index recommendation(s)`
      : "None";
  const classname =
    database.details?.stats?.num_index_recommendations > 0
      ? "index-recommendations-icon__exist"
      : "index-recommendations-icon__none";
  return (
    <div>
      <CircleFilled className={cx(classname)} />
      <span>{text}</span>
    </div>
  );
};

export const DatabaseNameCell = ({ database }: CellProps): JSX.Element => {
  const isCockroachCloud = useContext(CockroachCloudContext);
  const linkURL = isCockroachCloud
    ? `${location.pathname}/${database.name}`
    : EncodeDatabaseUri(database.name);
  let icon = <StackIcon className={cx("icon--s", "icon--primary")} />;

  if (database.error) {
    const titleList = [getNetworkErrorMessage(database.error)];
    icon = (
      <Tooltip
        overlayStyle={{ whiteSpace: "pre-line" }}
        placement="bottom"
        title={titleList.join("\n")}
      >
        <Caution className={cx("icon--s", "icon--warning")} />
      </Tooltip>
    );
  }
  return (
    <>
      <Link to={linkURL} className={cx("icon__container")}>
        {icon}
        {database.name}
      </Link>
    </>
  );
};
