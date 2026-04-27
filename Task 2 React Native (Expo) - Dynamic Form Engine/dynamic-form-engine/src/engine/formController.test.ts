import { describe, expect, it } from "vitest";
import { FormController } from "./formController";
import { FormSchema } from "../types/schema";

const schema: FormSchema = {
  id: "unit-form",
  fields: [
    {
      id: "firstName",
      label: "First Name",
      type: "text",
      validation: [
        { type: "required" },
        { type: "minLength", value: 2 },
      ],
    },
    {
      id: "role",
      label: "Role",
      type: "select",
      options: [
        { label: "Developer", value: "dev" },
        { label: "Manager", value: "manager" },
      ],
      validation: [{ type: "required" }],
    },
    {
      id: "badge",
      label: "Badge",
      type: "text",
      visibleWhen: { field: "role", equals: "manager" },
      validation: [{ type: "required" }],
    },
    {
      id: "startDate",
      label: "Start Date",
      type: "date",
      visibleWhen: { field: "role", equals: "manager" },
      validation: [
        { type: "required" },
        {
          type: "regex",
          value: "^\\d{4}-(0[1-9]|1[0-2])-(0[1-9]|[12]\\d|3[01])$",
          message: "Start Date must use YYYY-MM-DD format",
        },
      ],
    },
    {
      id: "email",
      label: "Email",
      type: "text",
      validation: [{ type: "custom", validator: "emailPolicy" }],
    },
  ],
};

describe("FormController", () => {
  it("tracks pristine to dirty transition", () => {
    const controller = new FormController({ schema });
    expect(controller.getState().status).toBe("pristine");

    const next = controller.setFieldValue("firstName", "A");
    expect(next.status).toBe("dirty");
    expect(next.dirtyFields.has("firstName")).toBe(true);
  });

  it("ignores validation on hidden fields", () => {
    const controller = new FormController({
      schema,
      initialValues: { firstName: "Tom", role: "dev" },
      validatorRegistry: {
        emailPolicy: (value) => (value ? null : "Email required"),
      },
    });
    const result = controller.validate();
    const badgeErrors = result.errors.filter((error) => error.fieldId === "badge");
    expect(badgeErrors.length).toBe(0);
  });

  it("evaluates custom validators from registry", () => {
    const controller = new FormController({
      schema,
      initialValues: {
        firstName: "Tom",
        role: "manager",
        badge: "B2",
        startDate: "2026-04-27",
        email: "bad",
      },
      validatorRegistry: {
        emailPolicy: (value) =>
          typeof value === "string" && value.endsWith("@company.com")
            ? null
            : "Use company email",
      },
    });
    const result = controller.validate();
    expect(result.errors.some((error) => error.message === "Use company email")).toBe(true);
  });

  it("moves to success when submit passes", async () => {
    const controller = new FormController({
      schema,
      initialValues: {
        firstName: "Tom",
        role: "manager",
        badge: "B2",
        startDate: "2026-04-27",
        email: "tom@company.com",
      },
      validatorRegistry: {
        emailPolicy: () => null,
      },
    });
    const result = await controller.submit(async () => true);
    expect(result.status).toBe("success");
  });

  it("rejects invalid manager start date format", () => {
    const controller = new FormController({
      schema,
      initialValues: {
        firstName: "Tom",
        role: "manager",
        badge: "B2",
        startDate: "111",
        email: "tom@company.com",
      },
      validatorRegistry: {
        emailPolicy: () => null,
      },
    });
    const result = controller.validate();
    expect(
      result.errors.some(
        (error) =>
          error.fieldId === "startDate" &&
          error.message === "Start Date must use YYYY-MM-DD format"
      )
    ).toBe(true);
  });
});
