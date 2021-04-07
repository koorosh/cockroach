import React from "react";
import { assert } from "chai";
import { mount } from "enzyme";
import { Action, Store } from "redux";
import { Provider } from "react-redux";
import { ConnectedRouter } from "connected-react-router";
import { cockroach } from "src/js/protos";
import { Network, NetworkProps } from "./index";
import { getNetworkPropsFixture } from "./network.fixture";
import { Latency } from "./latency";
import { AdminUIState, createAdminUIStore } from "src/redux/state";
import NodeLivenessStatus = cockroach.kv.kvserver.liveness.livenesspb.NodeLivenessStatus;

const getNetworkPageWrapper = (props: NetworkProps) => {
  const store: Store<AdminUIState, Action> = createAdminUIStore(props.history);
  return mount(
    <Provider store={store}>
      <ConnectedRouter history={props.history}>
        <Network {...props} />
      </ConnectedRouter>
    </Provider>,
  );
};

describe("Network page", () => {
  it("renders latency matrix when some nodes are dead", () => {
    const props = getNetworkPropsFixture();
    // set node 1 as a dead node
    props.nodesSummary.livenessStatusByNodeID["1"] =
      NodeLivenessStatus.NODE_STATUS_DEAD;
    const wrapper = getNetworkPageWrapper(props);
    assert.isTrue(wrapper.exists());

    const latencyWrapper = wrapper.find(Latency);
    latencyWrapper.exists();
  });

  it("renders latency matrix when activity dict for node has missed node ids", () => {
    const props = getNetworkPropsFixture();
    delete props.nodesSummary.nodeStatusByID["4"].activity["1"];
    const wrapper = getNetworkPageWrapper(props);
    assert.isTrue(wrapper.exists());
    const latencyWrapper = wrapper.find(Latency);
    latencyWrapper.exists();
  });

  it("renders latency matrix when activity dict for node has network activity records without 'latency' field", () => {
    const props = getNetworkPropsFixture();
    delete props.nodesSummary.nodeStatusByID["4"].activity["1"].latency;
    const wrapper = getNetworkPageWrapper(props);
    assert.isTrue(wrapper.exists());
    const latencyWrapper = wrapper.find(Latency);
    latencyWrapper.exists();
  });
});
