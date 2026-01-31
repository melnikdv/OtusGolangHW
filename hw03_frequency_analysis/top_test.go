package hw03frequencyanalysis

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Change to true if needed.
var taskWithAsteriskIsCompleted = true

var text = `Как видите, он  спускается  по  лестнице  вслед  за  своим
	другом   Кристофером   Робином,   головой   вниз,  пересчитывая
	ступеньки собственным затылком:  бум-бум-бум.  Другого  способа
	сходить  с  лестницы  он  пока  не  знает.  Иногда ему, правда,
		кажется, что можно бы найти какой-то другой способ, если бы  он
	только   мог   на  минутку  перестать  бумкать  и  как  следует
	сосредоточиться. Но увы - сосредоточиться-то ему и некогда.
		Как бы то ни было, вот он уже спустился  и  готов  с  вами
	познакомиться.
	- Винни-Пух. Очень приятно!
		Вас,  вероятно,  удивляет, почему его так странно зовут, а
	если вы знаете английский, то вы удивитесь еще больше.
		Это необыкновенное имя подарил ему Кристофер  Робин.  Надо
	вам  сказать,  что  когда-то Кристофер Робин был знаком с одним
	лебедем на пруду, которого он звал Пухом. Для лебедя  это  было
	очень   подходящее  имя,  потому  что  если  ты  зовешь  лебедя
	громко: "Пу-ух! Пу-ух!"- а он  не  откликается,  то  ты  всегда
	можешь  сделать вид, что ты просто понарошку стрелял; а если ты
	звал его тихо, то все подумают, что ты  просто  подул  себе  на
	нос.  Лебедь  потом  куда-то делся, а имя осталось, и Кристофер
	Робин решил отдать его своему медвежонку, чтобы оно не  пропало
	зря.
		А  Винни - так звали самую лучшую, самую добрую медведицу
	в  зоологическом  саду,  которую  очень-очень  любил  Кристофер
	Робин.  А  она  очень-очень  любила  его. Ее ли назвали Винни в
	честь Пуха, или Пуха назвали в ее честь - теперь уже никто  не
	знает,  даже папа Кристофера Робина. Когда-то он знал, а теперь
	забыл.
		Словом, теперь мишку зовут Винни-Пух, и вы знаете почему.
		Иногда Винни-Пух любит вечерком во что-нибудь поиграть,  а
	иногда,  особенно  когда  папа  дома,  он больше любит тихонько
	посидеть у огня и послушать какую-нибудь интересную сказку.
		В этот вечер...`

func TestTop10(t *testing.T) {
	t.Run("no words in empty string", func(t *testing.T) {
		require.Len(t, Top10(""), 0)
	})

	t.Run("positive test", func(t *testing.T) {
		if taskWithAsteriskIsCompleted {
			expected := []string{
				"а",         // 8
				"он",        // 8
				"и",         // 6
				"ты",        // 5
				"что",       // 5
				"в",         // 4
				"его",       // 4
				"если",      // 4
				"кристофер", // 4
				"не",        // 4
			}
			require.Equal(t, expected, Top10(text))
		} else {
			expected := []string{
				"он",        // 8
				"а",         // 6
				"и",         // 6
				"ты",        // 5
				"что",       // 5
				"-",         // 4
				"Кристофер", // 4
				"если",      // 4
				"не",        // 4
				"то",        // 4
			}
			require.Equal(t, expected, Top10(text))
		}
	})
}

func TestTop10SpecificText(t *testing.T) {
	t.Run("test with specific text - check word counts", func(t *testing.T) {
		result := Top10(text)

		// Проверяем, что возвращается 10 слов
		require.Len(t, result, 10)

		// Проверяем, что слова не пустые
		for _, word := range result {
			require.NotEmpty(t, word)
		}

		// Проверяем, что все слова из текста присутствуют в результатах
		// (или хотя бы некоторые ключевые слова)
		wordsInResult := make(map[string]bool)
		for _, word := range result {
			wordsInResult[word] = true
		}

		// Проверяем наличие ключевых слов
		require.True(t, wordsInResult["он"] || wordsInResult["а"] || wordsInResult["и"])
	})

	t.Run("test with specific text - verify frequency order", func(t *testing.T) {
		result := Top10(text)

		// Проверяем, что результат не пустой
		require.NotEmpty(t, result)

		// Проверяем, что первые слова имеют наибольшую частоту
		// Слово "а" должно быть первым (8 вхождений)
		require.Equal(t, "а", result[0])
	})

	t.Run("test with specific text - check special characters", func(t *testing.T) {
		result := Top10(text)

		// Проверяем, что специальные символы обрабатываются правильно
		// Проверяем, что нет пустых слов
		for _, word := range result {
			require.NotEqual(t, "", word)
		}
	})
}

func TestTop10WithAsterisk(t *testing.T) {
	t.Run("test case insensitive", func(t *testing.T) {
		result := Top10("Нога нога НОГА")
		require.Len(t, result, 1)
		require.Equal(t, "нога", result[0])
	})

	t.Run("test punctuation trimming", func(t *testing.T) {
		result := Top10("нога! нога, 'нога' нога")
		require.Len(t, result, 1)
		require.Equal(t, "нога", result[0])
	})

	t.Run("test different words with hyphen", func(t *testing.T) {
		result := Top10("какой-то какойто")
		// Должны быть два разных слова
		require.Len(t, result, 2)
		require.Contains(t, result, "какой-то")
		require.Contains(t, result, "какойто")
	})

	t.Run("test different compound words", func(t *testing.T) {
		result := Top10("dog,cat dog...cat dogcat")
		// Должны быть три разных слова
		require.Len(t, result, 3)
		require.Contains(t, result, "dog,cat")
		require.Contains(t, result, "dog...cat")
		require.Contains(t, result, "dogcat")
	})

	t.Run("test hyphen alone is not word", func(t *testing.T) {
		result := Top10("hello - world")
		// "-" не должно быть отдельным словом
		require.NotContains(t, result, "-")
		require.Contains(t, result, "hello")
		require.Contains(t, result, "world")
	})

	t.Run("test mixed case and punctuation", func(t *testing.T) {
		result := Top10("Hello, HELLO! Hello. 'hello'")
		require.Len(t, result, 1)
		require.Equal(t, "hello", result[0])
	})

	t.Run("test complex punctuation", func(t *testing.T) {
		result := Top10("word!!! ???word... word???")
		require.Len(t, result, 1)
		require.Equal(t, "word", result[0])
	})

	t.Run("test words with hyphens", func(t *testing.T) {
		result := Top10("hello-world test-word example-word")
		// Слова с тире должны сохраняться как есть
		require.Len(t, result, 3)
		require.Contains(t, result, "hello-world")
		require.Contains(t, result, "test-word")
		require.Contains(t, result, "example-word")
	})

	t.Run("test multiple hyphens as separate words", func(t *testing.T) {
		result := Top10("----- -- ---")
		// Больше одного тире считается словом
		require.NotEmpty(t, result)
		for _, word := range result {
			require.NotEqual(t, "", word)
		}
	})

	t.Run("test single hyphen is not word", func(t *testing.T) {
		result := Top10("word - another")
		// Одиночное тире не должно быть словом
		require.NotContains(t, result, "-")
		require.Contains(t, result, "word")
		require.Contains(t, result, "another")
	})
}
