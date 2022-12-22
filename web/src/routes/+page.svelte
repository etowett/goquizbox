<script>
	import Pagination from './Pagination.svelte';

  /** @type {import('./$types').PageData} */
	export let data;
</script>

<svelte:head>
	<title>Quizbox</title>
</svelte:head>

<h1>Question feed</h1>

<div>
{#if data.questions === null}
	<div>No questions are here ... yet.</div>
{:else}
	<div>
		{#each data.questions as question (question.id)}
			<div class="question">
        <h3><a href="/questions/{question.id}">{question.title}</a></h3>
        <small>
          By {question.user_id} asked {question.created_at}
          <a href="/questions/{question.id}/vote?act=up&kind=question">Upvote</a> |
          <a href="/questions/{question.id}/vote?act=down&kind=question">Downvote</a> |
        </small>
        <hr />
      </div>
		{/each}
    <Pagination pagination={data.pagination} />
	</div>
{/if}
</div>
