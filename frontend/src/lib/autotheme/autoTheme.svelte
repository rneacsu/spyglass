<script lang="ts" module>
  export let autoTheme = $state({
    theme: "light",
  })
</script>

<script lang="ts">
  import { onDestroy, onMount } from "svelte";

  function setTheme(theme: string) {
    document.documentElement.setAttribute("data-bs-theme", theme);
    autoTheme.theme = theme;
  }

  function getPreferredTheme() {
    return window.matchMedia("(prefers-color-scheme: dark)").matches
      ? "dark"
      : "light";
  }

  function onThemeChange(event: MediaQueryListEvent) {
    setTheme(event.matches ? "dark" : "light");
  }

  onMount(() => {
    setTheme(getPreferredTheme());

    window
      .matchMedia("(prefers-color-scheme: dark)")
      .addEventListener("change", onThemeChange);
  });

  onDestroy(() => {
    window
      .matchMedia("(prefers-color-scheme: dark)")
      .removeEventListener("change", onThemeChange);
  });
</script>
