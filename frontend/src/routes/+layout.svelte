<script lang="ts">
  import AutoTheme from "$lib/autotheme/autoTheme.svelte";

  let { children } = $props();

  import "../app.scss";
  import "bootstrap";
</script>

<AutoTheme />

{@render children()}
