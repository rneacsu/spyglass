<script lang="ts">
  import grpcClient from "$lib/grpc/client";
  import { LogError, LogInfo } from "$lib/wailsjs/runtime/runtime";
    import { ConnectError } from "@connectrpc/connect";

  let resultText = "Please enter your name above 👆";
  let name: string;
  let contextsMessage = "Loading contexts...";
  let contexts: string[] = [];

  async function greet() {
    LogInfo("Greeting " + name);
    try {
      resultText = (await (await grpcClient).sayHello({ name })).message;
    } catch (e) {
      LogError("Error greeting: " + e);
      resultText = "Error greeting: " + e;
    }
  }

  grpcClient.then(async (client) => {
    contexts = (await client.getContexts({})).contexts;
    contextsMessage = "";
  }).catch((e) => {
    contextsMessage = ConnectError.from(e).message;
  });

</script>

<div class="p-3">

<h1>Welcome to SpyGlass</h1>

<form class="mt-3" on:submit|preventDefault={greet}>
  <div class="mb-3">
    <label for="name" class="form-label">Name</label>
    <input type="text" bind:value={name} class="form-control" id="name" />
  </div>

  <button type="submit" class="btn btn-primary"><i class="bi bi-chat-fill"></i> Greet</button>
</form>

<div class="mt-3">{resultText}</div>

<h2 class="mt-3">Contexts</h2>
<ul class="mt-3">
  {contextsMessage}
  {#each contexts as context}
    <li>{context}</li>
  {/each}
</ul>

</div>
