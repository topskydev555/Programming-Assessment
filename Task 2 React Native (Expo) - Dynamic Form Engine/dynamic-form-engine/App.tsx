import { SafeAreaView, ScrollView, StyleSheet, Text } from "react-native";
import { StatusBar } from "expo-status-bar";
import { FormRenderer } from "./src/components/FormRenderer";
import {
  defaultSchema,
  sampleCustomValidators,
  sampleInitialValues,
} from "./src/sampleSchema";

export default function App() {
  return (
    <SafeAreaView style={styles.root}>
      <StatusBar style="auto" />
      <ScrollView contentContainerStyle={styles.content}>
        <Text style={styles.title}>Dynamic Form Engine</Text>
        <FormRenderer
          schema={defaultSchema}
          initialValues={sampleInitialValues}
          customValidators={sampleCustomValidators}
          onSubmit={async (values) => {
            await new Promise((resolve) => setTimeout(resolve, 400));
            return values;
          }}
        />
      </ScrollView>
    </SafeAreaView>
  );
}

const styles = StyleSheet.create({
  root: { flex: 1, backgroundColor: "#fff" },
  content: { padding: 16, gap: 12 },
  title: { fontSize: 22, fontWeight: "700" },
});
