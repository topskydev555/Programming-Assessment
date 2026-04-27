import React, { useState } from "react";
import { Button, Text, View } from "react-native";
import { mockApiClient, setMockFailuresBeforeSuccess } from "../api/mockApiClient";
import { useApiClient } from "../hooks/useApiClient";

type DemoResponse = {
  id: string;
  value: string;
  fetchedAt: string;
};

export function ApiClientDemoScreen() {
  const [requestKey, setRequestKey] = useState("https://demo.local/resource");
  const { state, execute, cancel } = useApiClient<DemoResponse>(mockApiClient, {
    cacheTtlMs: 4_000,
    maxAttempts: 3,
    baseDelayMs: 200,
  });

  const runProfile = async () => {
    setMockFailuresBeforeSuccess(1);
    await execute(requestKey);
  };

  const runDuplicateRequest = async () => {
    setMockFailuresBeforeSuccess(0);
    await Promise.all([execute(requestKey), execute(requestKey)]);
  };

  const currentState = state.error?.code === "CANCELLED"
    ? "cancelled"
    : state.loading
    ? "loading"
    : state.error
    ? "error"
    : state.data
    ? "success"
    : "idle";

  const lastSource = state.fromCache ? "cache" : state.data ? "network" : "none";
  const payloadTime = state.data?.fetchedAt ?? "none";

  return (
    <View style={{ padding: 16, gap: 10 }}>
      <Text style={{ fontSize: 30, fontWeight: "700", marginBottom: 12 }}>
        Task 3 API Hook Demo
      </Text>
      <Text>State: {currentState}</Text>
      <Text>Last source: {lastSource}</Text>
      <Text>Error: {state.error?.message ?? "none"}</Text>
      <Text>Payload time: {payloadTime}</Text>
      <Button title="LOAD PROFILE" onPress={runProfile} />
      <Button title="LOAD DUPLICATE REQUEST" onPress={runDuplicateRequest} />
      <Button title="CANCEL PENDING" onPress={cancel} />
      <Button
        title="CHANGE REQUEST KEY"
        onPress={() => setRequestKey(`${requestKey}?v=${Date.now()}`)}
      />
    </View>
  );
}
