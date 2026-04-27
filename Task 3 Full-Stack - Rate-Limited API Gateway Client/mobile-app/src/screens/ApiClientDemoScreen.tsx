import React, { useMemo } from "react";
import { Button, SafeAreaView, StyleSheet, Text, View } from "react-native";
import { MockApiClient } from "../api/mockApiClient";
import { useApiClient } from "../hooks/useApiClient";

type DemoPayload = {
  key: string;
  url: string;
  at: string;
  message: string;
};

export function ApiClientDemoScreen() {
  const client = useMemo(() => {
    const mock = new MockApiClient();
    mock.setTransientFailures("profile", 2);
    mock.setDelayMs("profile", 800);
    return mock;
  }, []);

  const { state, data, error, lastSource, execute, cancel } = useApiClient<DemoPayload>(client, {
    ttlMs: 10_000,
    retry: {
      maxAttempts: 3,
      baseDelayMs: 200,
    },
  });

  const request = {
    key: "profile",
    url: "https://gateway.local/profile",
    method: "GET" as const,
  };

  return (
    <SafeAreaView style={styles.root}>
      <Text style={styles.title}>Task 3 API Hook Demo</Text>
      <Text>State: {state}</Text>
      <Text>Last source: {lastSource}</Text>
      <Text>Error: {error ?? "none"}</Text>
      <Text>Payload time: {data?.at ?? "none"}</Text>

      <View style={styles.actions}>
        <Button title="Load profile" onPress={() => void execute(request)} />
      </View>
      <View style={styles.actions}>
        <Button title="Load duplicate request" onPress={() => { void execute(request); void execute(request); }} />
      </View>
      <View style={styles.actions}>
        <Button title="Cancel pending" onPress={() => cancel("profile")} />
      </View>
    </SafeAreaView>
  );
}

const styles = StyleSheet.create({
  root: {
    flex: 1,
    justifyContent: "center",
    alignItems: "center",
    gap: 8,
    padding: 16,
  },
  title: {
    fontSize: 22,
    marginBottom: 8,
    fontWeight: "600",
  },
  actions: {
    width: 260,
    marginTop: 8,
  },
});
