<script>
	import { page } from '$app/stores';
</script>

<header>
  <div class="d-flex flex-column flex-md-row align-items-center pb-3 mb-4 border-bottom">
    <a href="/" class="d-flex align-items-center text-dark text-decoration-none">
    <span class="fs-4">Quizbox</span>
    </a>

    <nav class="d-inline-flex mt-2 mt-md-0 ms-md-auto">
    <a class="me-3 py-2 text-dark text-decoration-none" href="#">Questions</a>
    <a class="me-3 py-2 text-dark text-decoration-none" href="#">Tags</a>
    {#if $page.data.user}
      <a class="me-3 py-2 text-dark text-decoration-none" href="/questions/new">Add Question</a>
      <a class="me-3 py-2 text-dark text-decoration-none" href="/profile/{$page.data.user.id}">Profile</a>
      <a class="me-3 py-2 text-dark text-decoration-none" href="/logout">Logout</a>
    {:else}
      <a class="me-3 py-2 text-dark text-decoration-none" href="/login">Login</a>
      <a class="me-3 py-2 text-dark text-decoration-none" href="/register">Signup</a>
    {/if}
    </nav>
  </div>
</header>
