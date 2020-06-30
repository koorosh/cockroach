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
import { Tooltip, Anchor } from "ui-modules/src/components";
import { nodeLivenessIssues, howItWork, capacityMetrics } from "src/util/docs";
import { LivenessStatus } from "ui-modules/src/core/redux/nodes";
import { AggregatedNodeStatus } from ".";

export const getStatusDescription = (status: LivenessStatus) => {
  switch (status) {
    case LivenessStatus.LIVE:
      return (
        <div className="tooltip__table--title">
          <p>
            {"This node is online and updating its "}
            <Anchor href={nodeLivenessIssues} target="_blank">
              liveness record
            </Anchor>
            .
          </p>
        </div>
      );
    case LivenessStatus.UNKNOWN:
    case LivenessStatus.UNAVAILABLE:
      return (
        <div className="tooltip__table--title">
          <p>
            {"This node has an "}
            <Anchor href={nodeLivenessIssues} target="_blank">
              unavailable liveness
            </Anchor>
            {" status."}
          </p>
        </div>
      );
    case LivenessStatus.DEAD:
      return (
        <div className="tooltip__table--title">
          <p>
            {"This node has not updated its "}
            <Anchor href={nodeLivenessIssues} target="_blank">
              liveness record
            </Anchor>
            {" for 5 minutes. CockroachDB "}
            <Anchor href={howItWork} target="_blank">
              automatically rebalances replicas
            </Anchor>
            {" from dead nodes to live nodes."}
          </p>
        </div>
      );
    case LivenessStatus.DECOMMISSIONING:
      return (
        <div className="tooltip__table--title">
          <p>
            {"This node is in the "}
            <Anchor href={howItWork} target="_blank">
              process of decommissioning
            </Anchor>
            {
              " , and may need time to transfer its data to other nodes. When finished, the node will appear below in the list of decommissioned nodes."
            }
          </p>
        </div>
      );
    default:
      return (
        "This node has not recently reported as being live. " +
        "It may not be functioning correctly, but no automatic action has yet been taken."
      );
  }
};

export const getNodeStatusDescription = (status: AggregatedNodeStatus) => {
  switch (status) {
    case AggregatedNodeStatus.LIVE:
      return (
        <div className="tooltip__table--title">
          <p>All nodes in this locality are live.</p>
        </div>
      );
    case AggregatedNodeStatus.WARNING:
      return (
        <div className="tooltip__table--title">
          <p>
            This locality has 1 or more <code>SUSPECT</code> or{" "}
            <code>DECOMMISSIONING</code> nodes.
          </p>
        </div>
      );
    case AggregatedNodeStatus.DEAD:
      return (
        <div className="tooltip__table--title">
          <p>
            This locality has 1 or more <code>DEAD</code> nodes.
          </p>
        </div>
      );
    default:
      return "This node is decommissioned and has been permanently removed from this cluster.";
  }
};

export const NodeCountTooltip: React.FC<React.PropsWithChildren<{}>> = ({
  children,
}) => (
  <Tooltip
    placement="bottom"
    title={
      <div className="tooltip__table--title">
        <p>Number of nodes in the locality.</p>
      </div>
    }
  >
    {children}
  </Tooltip>
);

export const UptimeTooltip: React.FC<React.PropsWithChildren<{}>> = ({
  children,
}) => (
  <Tooltip
    placement="bottom"
    title={
      <div className="tooltip__table--title">
        <p>Amount of time the node has been running.</p>
      </div>
    }
  >
    {children}
  </Tooltip>
);

export const ReplicasTooltip: React.FC<React.PropsWithChildren<{}>> = ({
  children,
}) => (
  <Tooltip
    placement="bottom"
    title={
      <div className="tooltip__table--title">
        <p>Number of replicas on the node or in the locality.</p>
      </div>
    }
  >
    {children}
  </Tooltip>
);

export const CapacityUseTooltip: React.FC<React.PropsWithChildren<{}>> = ({
  children,
}) => (
  <Tooltip
    placement="bottom"
    title={
      <div className="tooltip__table--title">
        <p>
          Percentage of usable disk space occupied by CockroachDB data at the
          locality or node.
        </p>
        <p>
          <Anchor href={capacityMetrics} target="_blank">
            How is this metric calculated?
          </Anchor>
        </p>
      </div>
    }
  >
    {children}
  </Tooltip>
);

export const MemoryUseTooltip: React.FC<React.PropsWithChildren<{}>> = ({
  children,
}) => (
  <Tooltip
    placement="bottom"
    title={
      <div className="tooltip__table--title">
        <p>
          Percentage of total memory at the locality or node in use by
          CockroachDB.
        </p>
      </div>
    }
  >
    {children}
  </Tooltip>
);

export const CPUsTooltip: React.FC<React.PropsWithChildren<{}>> = ({
  children,
}) => (
  <Tooltip
    placement="bottom"
    title={
      <div className="tooltip__table--title">
        <p>Number of vCPUs on the machine.</p>
      </div>
    }
  >
    {children}
  </Tooltip>
);

export const VersionTooltip: React.FC<React.PropsWithChildren<{}>> = ({
  children,
}) => (
  <Tooltip
  placement="bottom"
  title={
    <div className="tooltip__table--title">
      <p>
        Build tag of the CockroachDB version installed on the node.
      </p>
    </div>
  }
>
    {children}
  </Tooltip>
);

export const StatusTooltip: React.FC<React.PropsWithChildren<{}>> = ({
  children,
}) => (
  <Tooltip
  placement="bottom"
  title={
    <div className="tooltip__table--title">
      <p>
        Node status can be live, suspect, dead, decommissioning, or decommissioned. Hover over the status for each node to learn more.
      </p>
    </div>
  }
  >
    {children}
  </Tooltip>
);
