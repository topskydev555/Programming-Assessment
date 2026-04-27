import { FormSchema } from "./types/schema";
import { ValidatorRegistry } from "./engine/validation";

export const defaultSchema: FormSchema = {
  id: "profile-form",
  title: "Profile Setup",
  fields: [
    {
      id: "name",
      label: "Name",
      type: "text",
      placeholder: "Full name",
      validation: [
        { type: "required" },
        { type: "minLength", value: 3 },
        { type: "maxLength", value: 40 },
      ],
    },
    {
      id: "email",
      label: "Email",
      type: "text",
      placeholder: "name@example.com",
      validation: [
        { type: "required" },
        { type: "regex", value: "^[^\\s@]+@[^\\s@]+\\.[^\\s@]+$" },
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
      id: "startDate",
      label: "Start Date",
      type: "date",
      placeholder: "YYYY-MM-DD",
      validation: [
        { type: "required" },
        {
          type: "regex",
          value: "^\\d{4}-(0[1-9]|1[0-2])-(0[1-9]|[12]\\d|3[01])$",
          message: "Start Date must use YYYY-MM-DD format",
        },
      ],
      visibleWhen: { field: "role", equals: "manager" },
    },
    {
      id: "skills",
      label: "Skills",
      type: "multi-select",
      options: [
        { label: "React Native", value: "rn" },
        { label: "Go", value: "go" },
        { label: "System Design", value: "design" },
      ],
      validation: [{ type: "custom", validator: "atLeastOneSkill" }],
    },
  ],
};

export const sampleInitialValues = {
  role: "dev",
};

export const sampleCustomValidators: ValidatorRegistry = {
  atLeastOneSkill: (value) =>
    Array.isArray(value) && value.length > 0 ? null : "Select at least one skill",
};
