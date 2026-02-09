<script lang="ts">
  import Alerts from "$lib/alerts.svelte";
  import Navbar from "$lib/navbar.svelte";
  import Sidebar from "$lib/sidebar/sidebar.svelte";
  import Table from "$lib/table/table.svelte";

  let selected = $state({
    context: "",
    namespace: "",
    group: "",
    version: "",
    resource: "",
    namespaced: false,
  });
</script>

<div class="h-screen flex flex-col">
  <Navbar
    bind:context={selected.context}
    bind:namespace={selected.namespace}
    namespaced={selected.namespaced}
  />
  <Alerts />

  <div class="flex flex-row h-0 grow">
    <Sidebar
      context={selected.context}
      bind:group={selected.group}
      bind:version={selected.version}
      bind:resource={selected.resource}
      bind:namespaced={selected.namespaced}
    />

    <div class="py-2 w-0 grow relative">
      <div
        class="
          absolute inset-1/2 -translate-1/2 -z-100
          w-52 h-52 bg-base-300
          mask-[url('$lib/assets/k8s.svg')] mask-contain mask-no-repeat mask-center
        "
      ></div>
      <Table
        context={selected.context}
        namespace={selected.namespace}
        group={selected.group}
        version={selected.version}
        resource={selected.resource}
        namespaced={selected.namespaced}
      />
    </div>
  </div>
</div>
