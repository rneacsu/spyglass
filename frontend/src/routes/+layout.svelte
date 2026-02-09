<script lang="ts">
  let { children } = $props();

  import "../app.css";
</script>

<svelte:head>
    <title>SpyGlass</title>
</svelte:head>

{@render children?.()}
