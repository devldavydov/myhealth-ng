const editableNumberFormat = new Intl.NumberFormat("ru", {
  maximumFractionDigits: 1,
  useGrouping: false
});

export function formatEditableNumber(value: number): string {
  return editableNumberFormat.format(value);
}
