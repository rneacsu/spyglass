import type { ConfigColumns } from "datatables.net-bs5";

export let overrides = {
  "/v1::pods": {
    hiddenColumns: ["Nominated Node", "Readiness Gates"],
    columnOrder: ["Status", "Ready"],
    render: {
      "Status": (data: string) => {

        const statusMap = {
          "Running": "success",
          "Pending": "warning",
          "Succeeded": "info",
          "Failed": "danger",
          "Completed": "info",
          "CrashLoopBackOff": "danger",
        } as { [key: string]: string };

        return `<span class="badge bg-${statusMap[data] ?? "secondary"}">${data}</span>`;
      }
    }
  }
} as { [key: string]: {
  hiddenColumns: string[]
  render: { [key: string]: ConfigColumns['render'] }
} };
