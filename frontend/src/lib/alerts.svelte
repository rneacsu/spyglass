<script lang="ts" module>
  export class Alert {
    id: number;
    message: string;
    type: AlertType;
    timeout: number;
    timeoutId: number | null = null;

    static autoId = 0;

    constructor(message: string, type: AlertType = "default", timeout = 5000) {
      this.message = message;
      this.type = type;
      this.id = Alert.autoId++;
      this.timeout = timeout;
    }

    getAlertClass(): string {
      switch (this.type) {
        case "success":
          return "alert-success";
        case "default":
        case "info":
          return "alert-info";
        case "warning":
          return "alert-warning";
        case "error":
          return "alert-error";
      }
    }
  }
  export type AlertType = "default" | "success" | "info" | "warning" | "error";

  let alerts: Alert[] = $state([]);

  function addAlert(alert: Alert) {
    alerts.unshift(alert);

    if (alert.timeout > 0) {
      alert.timeoutId = window.setTimeout(() => {
        removeAlert(alert.id);
      }, alert.timeout);
    }
  }

  function removeAlert(id: number) {
    const alert = alerts.find((a) => a.id === id);
    if (alert && alert.timeoutId) {
      clearTimeout(alert.timeoutId);
    }
    alerts.splice(
      alerts.findIndex((a) => a.id === id),
      1,
    );
  }

  export function ShowAlert(type: AlertType, message: string, timeout = 5000) {
    addAlert(new Alert(message, type, timeout));
  }
</script>

<script lang="ts">
  import { X } from "@lucide/svelte";
  import { fade } from "svelte/transition";
</script>

<div class="toast flex-col-reverse z-10 mask-t-from-60">
  {#each alerts as alert (alert.id)}
    <div
      class="alert {alert.getAlertClass()} max-w-160 max-h-86"
      role="alert"
      transition:fade
    >
      <span class="text-ellipsis line-clamp-4">
        {alert.message}
      </span>
      <button
        class="btn btn-square btn-ghost btn-sm"
        onclick={() => {
          removeAlert(alert.id);
        }}
      >
        <X />
      </button>
    </div>
  {/each}
</div>
