package seed

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

type task struct {
	DurationMin int
	Category    string
	Title       string
	Detail      string
}

type day struct {
	Name  string
	Order int
	Focus string
	Tasks []task
}

type week struct {
	Number int
	Title  string
	Theme  string
	Days   []day
}

type piece struct {
	Title      string
	Composer   string
	Category   string
	Difficulty int
	Phase      int
	XPReward   int
	IMSLPUrl   string
	SortOrder  int
}

func Run(pool *pgxpool.Pool) error {
	ctx := context.Background()

	var phaseCount int
	err := pool.QueryRow(ctx, `SELECT COUNT(*) FROM curriculum_phases`).Scan(&phaseCount)
	if err != nil {
		return fmt.Errorf("seed check: %w", err)
	}
	if phaseCount > 0 {
		log.Println("seed: already seeded, skipping")
		return nil
	}

	log.Println("seed: running...")

	var phaseID int
	err = pool.QueryRow(ctx,
		`INSERT INTO curriculum_phases (phase_number, title, description)
		 VALUES (1, 'Fundamentos', 'Base técnica, repertório inicial e teoria musical elementar')
		 RETURNING id`,
	).Scan(&phaseID)
	if err != nil {
		return fmt.Errorf("insert phase: %w", err)
	}

	for _, w := range weeks() {
		var weekID int
		err = pool.QueryRow(ctx,
			`INSERT INTO curriculum_weeks (phase_id, week_number, title, theme)
			 VALUES ($1, $2, $3, $4) RETURNING id`,
			phaseID, w.Number, w.Title, w.Theme,
		).Scan(&weekID)
		if err != nil {
			return fmt.Errorf("insert week %d: %w", w.Number, err)
		}

		for _, d := range w.Days {
			var dayID int
			err = pool.QueryRow(ctx,
				`INSERT INTO curriculum_days (week_id, day_name, day_order, focus)
				 VALUES ($1, $2, $3, $4) RETURNING id`,
				weekID, d.Name, d.Order, d.Focus,
			).Scan(&dayID)
			if err != nil {
				return fmt.Errorf("insert day %s w%d: %w", d.Name, w.Number, err)
			}

			for i, t := range d.Tasks {
				detail := &t.Detail
				if t.Detail == "" {
					detail = nil
				}
				_, err = pool.Exec(ctx,
					`INSERT INTO curriculum_tasks (day_id, task_order, duration_min, category, title, detail)
					 VALUES ($1, $2, $3, $4, $5, $6)`,
					dayID, i+1, t.DurationMin, t.Category, t.Title, detail,
				)
				if err != nil {
					return fmt.Errorf("insert task %s day %s w%d: %w", t.Title, d.Name, w.Number, err)
				}
			}
		}
	}

	for _, p := range pieces() {
		imslp := &p.IMSLPUrl
		if p.IMSLPUrl == "" {
			imslp = nil
		}
		_, err = pool.Exec(ctx,
			`INSERT INTO pieces (title, composer, category, difficulty, phase, xp_reward, imslp_url, sort_order)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
			p.Title, p.Composer, p.Category, p.Difficulty, p.Phase, p.XPReward, imslp, p.SortOrder,
		)
		if err != nil {
			return fmt.Errorf("insert piece %s: %w", p.Title, err)
		}
	}

	log.Println("seed: done")
	return nil
}

func weeks() []week {
	return []week{
		{
			Number: 1,
			Title:  "O Despertar do Pianista",
			Theme:  "Aquecimento, escalas e primeiros contatos com novos repertórios",
			Days: []day{
				{Name: "Segunda-feira", Order: 1, Focus: "Técnica + Bach", Tasks: []task{
					{15, "tecnica", "Escalas de Dó maior", "Mãos separadas → juntas, 2 oitavas, BPM 60"},
					{20, "tecnica", "Hanon n°1", "BPM 60, legato, sem tensão nos ombros"},
					{25, "tecnica", "Czerny Op.261 n°4 (novo)", "Ler MDir, depois MEsq, depois juntas — devagar"},
					{30, "classico", "Bach Prelúdio n°1 — polimento", "3x completo. Foco em dinâmica suave e contínua"},
					{20, "anime", "Pokémon Main Theme — melodia", "Apenas a melodia MDir, aprender as notas"},
					{10, "teoria", "Leitura à primeira vista", "Peça simples desconhecida — leia sem parar"},
				}},
				{Name: "Terça-feira", Order: 2, Focus: "Lirismo + Melodia", Tasks: []task{
					{15, "tecnica", "Escalas de Sol maior", "Mesma dinâmica de segunda, BPM 60"},
					{20, "tecnica", "Hanon n°2", "BPM 60, staccato desta vez"},
					{25, "classico", "Burgmüller Op.100 n°3 \"Pastoral\"", "Ler a peça toda, identificar a forma A-B-A"},
					{30, "classico", "Czerny Op.599 n°19 — polimento", "3 passagens completas, corrigindo enganos"},
					{25, "pop", "Titanic — My Heart Will Go On", "Apenas a melodia da introdução, MDir, BPM lento"},
					{5, "teoria", "Teoria: escala maior e tons/semitons", "Escrever escala de Dó e Sol no caderno"},
				}},
				{Name: "Quarta-feira", Order: 3, Focus: "Técnica + Clássico", Tasks: []task{
					{15, "tecnica", "Escalas de Ré maior", "BPM 60 → 70 se confortável"},
					{20, "tecnica", "Hanon n°3", "BPM 60, alternando legato e staccato"},
					{25, "tecnica", "Czerny Op.261 n°5 (novo)", "Ler mãos separadas, identificar o padrão"},
					{35, "classico", "Bach Prelúdio n°1 — análise", "MEsq: identifique cada acorde. MDir: cante a linha"},
					{15, "pop", "Titanic — melodia + nota do baixo", "Adicione a nota do baixo com MEsq"},
					{10, "teoria", "Sight-reading livre", "Escolha um Burgmüller que não estudou e tente ler"},
				}},
				{Name: "Quinta-feira", Order: 4, Focus: "Anime + Exploração", Tasks: []task{
					{15, "tecnica", "Escalas de Lá maior", "BPM 60, atenção ao dedilhado (polegar no 3° grau)"},
					{20, "tecnica", "Hanon n°4", "BPM 60"},
					{25, "classico", "Burgmüller Op.100 n°4 \"Étude\"", "Atenção às colcheias em staccato na MDir"},
					{25, "classico", "Bach Prelúdio n°2 — introdução", "Apenas os 4 primeiros compassos, mãos separadas"},
					{25, "anime", "Naruto — Sadness and Sorrow", "Ouvir 2x, depois tentar a melodia principal na MDir"},
					{10, "teoria", "Teoria: tônica, dominante e subdominante", "Toque I, IV, V em Dó maior e identifique o som"},
				}},
				{Name: "Sexta-feira", Order: 5, Focus: "Revisão + Rock", Tasks: []task{
					{15, "tecnica", "Revisão: todas as escalas da semana", "Dó, Sol, Ré, Lá — 1x cada, BPM 70"},
					{20, "tecnica", "Revisão Hanon n°1–4", "1x cada, foco em qualidade, não velocidade"},
					{30, "classico", "Mini-recital: Bach + Burgmüller + Czerny", "Toque cada peça 1x sem parar"},
					{25, "rock", "Imagine — John Lennon", "Ouvir a gravação, aprender a melodia do verso (MDir)"},
					{20, "anime", "Pokémon + Naruto — revisão", "Revisão das melodias. Toque cada 2x"},
					{10, "teoria", "Auto-avaliação da semana", "O que melhorou? O que precisa de mais atenção?"},
				}},
				{Name: "Sábado", Order: 6, Focus: "Passagem Completa + Teoria", Tasks: []task{
					{20, "tecnica", "Escalas maiores: Dó, Sol, Ré, Lá, Mi", "BPM 70, dedilhado correto, sem hesitar"},
					{25, "classico", "Bach Prelúdio n°1 — performance", "Toque como se fosse para um professor. Grave se possível"},
					{25, "classico", "Burgmüller Op.100 n°2 e n°3", "Passagem das duas peças: dominada + nova"},
					{20, "pop", "River Flows in You — Yiruma (exploração)", "Ouvir, tentar a abertura de 4 compassos"},
					{30, "teoria", "Teoria: leitura de partituras + ritmo", "Figuras rítmicas: identificar nas páginas do Czerny"},
				}},
			},
		},
		{
			Number: 2,
			Title:  "Construindo Alicerces",
			Theme:  "Aumentar BPM do Hanon, aprofundar Bach, introduzir Beauty and the Beast",
			Days: []day{
				{Name: "Segunda-feira", Order: 1, Focus: "Técnica + Bach", Tasks: []task{
					{15, "tecnica", "Escalas: Dó e Sol (BPM 70)", "Mais fluente que na semana passada"},
					{20, "tecnica", "Hanon n°1 e n°2 (BPM 70)", ""},
					{25, "tecnica", "Czerny Op.261 n°6 (novo)", "Mãos separadas + juntas em câmera lenta"},
					{30, "classico", "Bach Prelúdio n°1 — memorização", "Tente os primeiros 12 compassos sem partitura"},
					{20, "pop", "Beauty and the Beast — melodia", "Melodia principal, MDir, BPM 50"},
					{10, "teoria", "Sight-reading: Burgmüller Op.100 n°5", "Leia pela primeira vez sem tocar antes"},
				}},
				{Name: "Terça-feira", Order: 2, Focus: "Lirismo + Pop", Tasks: []task{
					{15, "tecnica", "Escalas: Ré e Lá (BPM 70)", ""},
					{20, "tecnica", "Hanon n°3 e n°4 (BPM 70)", "Staccato e legato alternados"},
					{30, "classico", "Burgmüller Op.100 n°3 — consolidação", "Mãos juntas, trabalhar dinâmica piano/forte"},
					{25, "pop", "Titanic — melodia + bass notes", "Adicionar as notas do baixo (MEsq simples)"},
					{20, "anime", "Naruto: Sadness and Sorrow — MDir completa", "Continuar a aprender a melodia completa"},
					{10, "teoria", "Teoria: tons e semitons cromáticos", "Identificar onde estão os semitons na escala de Dó"},
				}},
				{Name: "Quarta-feira", Order: 3, Focus: "Técnica + Clássico", Tasks: []task{
					{15, "tecnica", "Escalas: Mi e Si (novas)", "BPM 60, atenção ao dedilhado certo"},
					{20, "tecnica", "Hanon n°5 (novo)", "BPM 60"},
					{25, "tecnica", "Czerny Op.261 n°7 (novo)", ""},
					{35, "classico", "Bach Prelúdio n°2 — 8 primeiros compassos", "Mãos separadas, depois juntas. Cuidado com o ritmo"},
					{15, "pop", "Beauty and the Beast — MDir + acorde MEsq", "Uma nota do acorde na MEsq + melodia na MDir"},
					{10, "teoria", "Sight-reading", ""},
				}},
				{Name: "Quinta-feira", Order: 4, Focus: "Anime + Bach", Tasks: []task{
					{15, "tecnica", "Escalas da semana (revisão)", ""},
					{20, "tecnica", "Hanon n°1–5 (1x cada)", "BPM 70 para n°1–4, BPM 60 para n°5"},
					{25, "classico", "Burgmüller Op.100 n°4 — consolidação", "Mãos juntas, trabalhar staccato preciso"},
					{25, "classico", "Bach Prelúdio n°2 — avançar 4 compassos", ""},
					{25, "anime", "Naruto: Sadness and Sorrow — MEsq (baixo)", "Aprender o acompanhamento da MEsq isolado"},
					{10, "teoria", "Teoria: compasso 4/4 e 3/4", "Identificar nos Prelúdios de Bach os pulsos"},
				}},
				{Name: "Sexta-feira", Order: 5, Focus: "Revisão + Rock", Tasks: []task{
					{15, "tecnica", "Revisão escalas da semana", ""},
					{20, "tecnica", "Revisão Hanon n°1–5", ""},
					{30, "classico", "Mini-recital: Bach + Burgmüller", "Prelúdio n°1 e n°2 (parcial), Burgmüller n°2, 3, 4"},
					{25, "rock", "Imagine — verso + refrão", "Adicionar o arpejo (MEsq) enquanto toca a melodia"},
					{20, "pop", "Revisão: Titanic + Beauty and the Beast", ""},
					{10, "teoria", "Auto-avaliação", "Qual peça melhorou mais? Qual precisa de mais atenção?"},
				}},
				{Name: "Sábado", Order: 6, Focus: "Teoria + Passagem", Tasks: []task{
					{20, "tecnica", "Escalas: todas da semana + revisão", "BPM 70–80"},
					{30, "classico", "Bach Prelúdio n°1 — quase memorizado", "Foco total na qualidade sonora e nas dinâmicas"},
					{25, "anime", "Spirited Away — One Summer's Day", "Ouvir Joe Hisaishi, tentar os primeiros 4 compassos"},
					{25, "rock", "Imagine — peça completa, BPM 60", "Mãos juntas, lento"},
					{20, "teoria", "Teoria: leitura de notas na clave de fá", "Exercícios básicos para a mão esquerda"},
				}},
			},
		},
		{
			Number: 3,
			Title:  "Velocidade e Expressão",
			Theme:  "Aumentar BPM do Hanon, trabalhar dinâmica, introduzir Dragon Ball",
			Days: []day{
				{Name: "Segunda-feira", Order: 1, Focus: "Técnica: velocidade", Tasks: []task{
					{15, "tecnica", "Escalas Dó–Si (BPM 80)", "Todas as maiores vistas até agora"},
					{25, "tecnica", "Hanon n°1–5 (BPM 80)", "Meta desta semana: BPM 80 limpo"},
					{20, "tecnica", "Czerny Op.261 n°8 e n°9", "Novos. Ler devagar, mãos separadas"},
					{30, "classico", "Bach Prelúdio n°1 — memorizado", "Toque sem partitura, 3 vezes"},
					{20, "anime", "Dragon Ball Z — Cha-La Head-Cha-La", "Ouvir, identificar a melodia, tentar as primeiras 4 barras"},
					{10, "teoria", "Sight-reading: peça nova do Burgmüller", ""},
				}},
				{Name: "Terça-feira", Order: 2, Focus: "Expressão Musical", Tasks: []task{
					{15, "tecnica", "Escalas BPM 80", ""},
					{20, "tecnica", "Hanon n°6 (novo)", "BPM 60"},
					{30, "classico", "Burgmüller Op.100 n°5 \"Innocence\"", "Peça lírica — foco no legato e fraseado"},
					{25, "classico", "Bach Prelúdio n°2 — metade", "Mãos juntas, metade da peça"},
					{20, "pop", "Beauty and the Beast — mãos juntas BPM 50", "Primeira tentativa mãos juntas"},
					{10, "teoria", "Dinâmica: p, mp, mf, f, ff", "Identificar nos Prelúdios e marcar na partitura"},
				}},
				{Name: "Quarta-feira", Order: 3, Focus: "Técnica + Clássico", Tasks: []task{
					{15, "tecnica", "Escalas BPM 80", ""},
					{20, "tecnica", "Hanon n°7 (novo)", ""},
					{25, "tecnica", "Czerny Op.261 n°10 (novo)", ""},
					{35, "classico", "Bach Prelúdio n°2 — passagem completa", "Mãos separadas na seção difícil, juntas no restante"},
					{15, "anime", "Dragon Ball — melodia completa MDir", ""},
					{10, "teoria", "Sight-reading", ""},
				}},
				{Name: "Quinta-feira", Order: 4, Focus: "Anime + Pop", Tasks: []task{
					{15, "tecnica", "Revisão escalas", ""},
					{20, "tecnica", "Hanon n°1–7 (revisão rápida)", "1x cada BPM 80"},
					{25, "classico", "Burgmüller n°5 — consolidação", ""},
					{25, "anime", "Naruto: Sadness and Sorrow — mãos juntas", "BPM 50, lento. Primeira tentativa mãos juntas"},
					{25, "pop", "Titanic — arranjo básico completo", "Melodia + baixo simples, do início ao fim"},
					{10, "teoria", "Articulação: legato vs. staccato", "Toque Burgmüller n°4 inteiro staccato, depois legato"},
				}},
				{Name: "Sexta-feira", Order: 5, Focus: "Revisão + Exploração", Tasks: []task{
					{15, "tecnica", "Escalas BPM 80", ""},
					{20, "tecnica", "Hanon n°1–7", ""},
					{30, "classico", "Mini-recital de revisão", "Todas as peças clássicas em ordem"},
					{25, "rock", "Piano Man — Billy Joel (melodia)", "Aprender a melodia do refrão"},
					{20, "anime", "Dragon Ball + Pokémon revisão", ""},
					{10, "teoria", "Auto-avaliação", ""},
				}},
				{Name: "Sábado", Order: 6, Focus: "Prática Livre + Teoria", Tasks: []task{
					{20, "tecnica", "Escala cromática (nova)", "Dó a Dó, 1 oitava, mãos separadas"},
					{30, "classico", "Bach Prelúdio n°1 — performance com dinâmicas", "Adicione crescendo e decrescendo intencionais"},
					{30, "pop", "River Flows in You — continuar", "Aprender até o compasso 16"},
					{20, "rock", "Piano Man — verso + refrão", ""},
					{20, "teoria", "Forma musical: A-B-A, tema, variação", "Análise dos Burgmüller estudados"},
				}},
			},
		},
		{
			Number: 4,
			Title:  "Consolidação do Mês 1",
			Theme:  "Revisar tudo do mês, subir BPM e preparar mini-recital pessoal",
			Days: []day{
				{Name: "Segunda-feira", Order: 1, Focus: "Consolidação Técnica", Tasks: []task{
					{15, "tecnica", "Todas as escalas do mês (BPM 80–90)", ""},
					{25, "tecnica", "Hanon n°1–7 (BPM 80)", "Igualdade de dedos, sem \"solavancos\""},
					{25, "tecnica", "Czerny Op.261 n°1–10 revisão", "2 peças por dia"},
					{35, "classico", "Bach Prelúdio n°1 + n°2", "Ambos do início ao fim, sem parar"},
					{10, "anime", "Spirited Away — continuar", "Avançar mais 4 compassos"},
					{10, "teoria", "Arpejos básicos em Dó maior", ""},
				}},
				{Name: "Terça-feira", Order: 2, Focus: "Pop e Anime", Tasks: []task{
					{15, "tecnica", "Escalas BPM 80", ""},
					{20, "tecnica", "Hanon n°8 (novo)", ""},
					{30, "pop", "Titanic + Beauty and the Beast — polimento", "Foco nas transições e fluência"},
					{30, "anime", "Naruto: Sadness and Sorrow — polimento", "Mãos juntas BPM 60, adicionar expressão"},
					{15, "rock", "Imagine — performance expressiva", ""},
					{10, "teoria", "Sight-reading", ""},
				}},
				{Name: "Quarta-feira", Order: 3, Focus: "Profundidade no Clássico", Tasks: []task{
					{15, "tecnica", "Escalas BPM 90", "Dó, Sol, Ré, Lá — tente BPM 90"},
					{20, "tecnica", "Hanon n°9 (novo)", ""},
					{25, "classico", "Burgmüller Op.100 n°6 \"Progresso\"", "Leitura inicial"},
					{35, "classico", "Bach Prelúdio n°2 — polimento", "Trabalhar dinâmicas e fraseado"},
					{15, "pop", "Harry Potter — Hedwig's Theme", "Ouvir, tentar os primeiros 4 compassos"},
					{10, "teoria", "Semicolcheia e ponto de aumento", ""},
				}},
				{Name: "Quinta-feira", Order: 4, Focus: "Rock + Anime", Tasks: []task{
					{15, "tecnica", "Escalas revisão", ""},
					{20, "tecnica", "Hanon n°10 (novo)", ""},
					{30, "rock", "Piano Man — peça completa BPM 70", "Verso, pré-refrão e refrão, mãos juntas"},
					{30, "anime", "Dragon Ball Z — mãos juntas", "BPM 55, primeira tentativa mãos juntas"},
					{15, "pop", "Harry Potter — avançar", "Mais 4 compassos"},
					{10, "teoria", "Acorde de 7ª dominante (introdução)", "Toque G7 em Dó e ouça a tensão-resolução"},
				}},
				{Name: "Sexta-feira", Order: 5, Focus: "Mini-Recital do Mês", Tasks: []task{
					{15, "tecnica", "Aquecimento: Hanon n°1–3 (BPM 80)", ""},
					{25, "classico", "Mini-recital clássico", "Bach Prelúdio n°1, n°2 | Burgmüller n°2, 3, 4, 5"},
					{30, "pop", "Mini-recital pop/cinema", "Titanic, Beauty and the Beast, Harry Potter (parcial)"},
					{30, "anime", "Mini-recital anime/rock", "Pokémon, Naruto, Dragon Ball (parcial), Imagine"},
					{20, "teoria", "Avaliação do mês 1", "O que domino? O que precisa de mais trabalho?"},
				}},
				{Name: "Sábado", Order: 6, Focus: "Revisão e Planejamento", Tasks: []task{
					{30, "tecnica", "Escalas + Hanon: revisão completa do mês", "Todas as escalas maiores, Hanon n°1–10"},
					{30, "classico", "Bach Prelúdios n°1 e n°2 — gravação", "Grave com o celular. Escute criticamente"},
					{30, "pop", "Escolha livre: peça favorita", "Aprofunde a peça que mais gostou"},
					{30, "teoria", "Revisão completa do mês 1", "Escalas, acordes, ritmo, articulação, dinâmica"},
				}},
			},
		},
		{
			Number: 5,
			Title:  "Arpejos e Inventividade",
			Theme:  "Introduzir arpejos, Bach Invenção n°1 (leitura), ampliar repertório pop",
			Days: []day{
				{Name: "Segunda-feira", Order: 1, Focus: "Arpejos + Hanon", Tasks: []task{
					{15, "tecnica", "Arpejos em Dó maior", "Mão direita, 1 oitava, BPM 60"},
					{25, "tecnica", "Hanon n°11 (novo)", "BPM 60"},
					{25, "tecnica", "Czerny Op.261 n°11–12", ""},
					{30, "classico", "Bach Invenção n°1 BWV 772 — leitura", "Só a voz superior (MDir), identificar o motivo"},
					{15, "pop", "Circle of Life — Rei Leão (melodia)", "Aprender a melodia principal"},
					{10, "teoria", "Arpejos e acordes quebrados", ""},
				}},
				{Name: "Terça-feira", Order: 2, Focus: "Pop + Clássico", Tasks: []task{
					{15, "tecnica", "Arpejos: Sol e Ré", ""},
					{20, "tecnica", "Hanon n°12", ""},
					{30, "classico", "Burgmüller Op.100 n°7 e n°8", "Leitura e prática mãos separadas"},
					{30, "pop", "Titanic + Beauty and the Beast polimento", "BPM mais próximo do real"},
					{15, "anime", "Spirited Away — avançar", ""},
					{10, "teoria", "Sight-reading", ""},
				}},
				{Name: "Quarta-feira", Order: 3, Focus: "Bach Invenção", Tasks: []task{
					{15, "tecnica", "Arpejos BPM 70", ""},
					{20, "tecnica", "Hanon n°13", ""},
					{25, "tecnica", "Czerny Op.261 n°13–14", ""},
					{35, "classico", "Bach Invenção n°1 — ambas as vozes separadas", "MEsq isolada, depois MDir, sem juntar ainda"},
					{15, "pop", "Circle of Life — avançar", ""},
					{10, "teoria", "Contraponto: o que é imitação?", "Identificar na Invenção n°1"},
				}},
				{Name: "Quinta-feira", Order: 4, Focus: "Anime Avançado", Tasks: []task{
					{15, "tecnica", "Arpejos revisão", ""},
					{20, "tecnica", "Hanon n°11–13 revisão", ""},
					{30, "classico", "Burgmüller n°9 e n°10", ""},
					{30, "anime", "Naruto: Sadness and Sorrow — polimento expressivo", "Dinâmicas, crescendos, rubato"},
					{20, "rock", "Piano Man — polimento", ""},
					{10, "teoria", "Escala menor natural de Lá", ""},
				}},
				{Name: "Sexta-feira", Order: 5, Focus: "Revisão Geral", Tasks: []task{
					{15, "tecnica", "Revisão escalas + arpejos", ""},
					{20, "tecnica", "Hanon n°1–13 resumo", "2 min cada"},
					{30, "classico", "Mini-recital clássico", ""},
					{30, "pop", "Mini-recital pop/anime/rock", ""},
					{15, "teoria", "Sight-reading + auto-avaliação", ""},
				}},
				{Name: "Sábado", Order: 6, Focus: "Teoria + Livre", Tasks: []task{
					{30, "tecnica", "Escalas menores: Lá e Mi (novo)", ""},
					{30, "classico", "Bach Invenção n°1 — tentar mãos juntas 8 compassos", ""},
					{30, "pop", "Peça livre: escolha o que mais gostou", ""},
					{30, "teoria", "Escalas menores harmônica e melódica", ""},
				}},
			},
		},
		{
			Number: 6,
			Title:  "Polifonia e Independência",
			Theme:  "Independência das mãos, Bach Invenção n°1 mãos juntas, Coldplay",
			Days: []day{
				{Name: "Segunda-feira", Order: 1, Focus: "Polifonia", Tasks: []task{
					{15, "tecnica", "Escalas menores (Lá, Mi) BPM 60", ""},
					{20, "tecnica", "Hanon n°14–15", ""},
					{25, "tecnica", "Czerny Op.261 n°15–16", ""},
					{35, "classico", "Bach Invenção n°1 — mãos juntas completo BPM 50", "Paciência! A independência vem com tempo"},
					{15, "rock", "Clocks — Coldplay (arpejo MEsq)", "Aprender o ostinato de arpejo da mão esquerda"},
					{10, "teoria", "Intervalos: terça, quinta, oitava", ""},
				}},
				{Name: "Terça-feira", Order: 2, Focus: "Melodia e Acompanhamento", Tasks: []task{
					{15, "tecnica", "Escalas menores BPM 70", ""},
					{20, "tecnica", "Hanon n°16", ""},
					{30, "classico", "Burgmüller n°11 \"O Cavalo-Árabe\"", "Peça característica com muita energia"},
					{30, "pop", "Harry Potter — peça completa BPM 55", "Mãos juntas"},
					{20, "rock", "Clocks — melodia + arpejo", ""},
					{5, "teoria", "Sight-reading", ""},
				}},
				{Name: "Quarta-feira", Order: 3, Focus: "Velocidade", Tasks: []task{
					{15, "tecnica", "Hanon n°1–5 (BPM 90)", "Meta: BPM 90 esta semana"},
					{20, "tecnica", "Hanon n°17", ""},
					{25, "tecnica", "Czerny Op.261 n°17–18", ""},
					{35, "classico", "Bach Invenção n°1 — BPM 60", "Aumentar um pouco o andamento"},
					{15, "anime", "Dragon Ball — polimento mãos juntas", ""},
					{10, "teoria", "Ornamentos — appogiatura", ""},
				}},
				{Name: "Quinta-feira", Order: 4, Focus: "Anime + Expressão", Tasks: []task{
					{15, "tecnica", "Arpejos todos (BPM 80)", ""},
					{20, "tecnica", "Hanon n°18", ""},
					{30, "classico", "Burgmüller n°12 e n°13", ""},
					{30, "anime", "Spirited Away — mãos juntas BPM 50", "Muito expressivo"},
					{20, "pop", "Circle of Life — avançar para seção B", ""},
					{10, "teoria", "Escalas menores de Ré e Sol", ""},
				}},
				{Name: "Sexta-feira", Order: 5, Focus: "Revisão", Tasks: []task{
					{15, "tecnica", "Revisão técnica semanal", ""},
					{20, "tecnica", "Hanon n°1–18 resumo", ""},
					{30, "classico", "Mini-recital", ""},
					{30, "pop", "Recital: Titanic, Harry Potter, Circle of Life", ""},
					{15, "teoria", "Auto-avaliação + anotações de progresso", ""},
				}},
				{Name: "Sábado", Order: 6, Focus: "Livre + Teoria", Tasks: []task{
					{25, "tecnica", "Escala cromática (2 oitavas)", ""},
					{30, "classico", "Bach Invenção n°1 — performance", "Do início ao fim sem parar"},
					{35, "rock", "Clocks (Coldplay) — peça completa", ""},
					{30, "teoria", "Introdução ao pedal de sustain", "Quando usar, quando não usar"},
				}},
			},
		},
		{
			Number: 7,
			Title:  "Mãos Fortes, Músico Mais Maduro",
			Theme:  "Clementi Sonatina (leitura), escalas em oitavas, Pirates of the Caribbean",
			Days: []day{
				{Name: "Segunda-feira", Order: 1, Focus: "Sonatina Clementi", Tasks: []task{
					{15, "tecnica", "Escalas maiores BPM 90", ""},
					{20, "tecnica", "Hanon n°19–20", ""},
					{25, "tecnica", "Czerny Op.261 n°19–20", ""},
					{35, "classico", "Clementi Sonatina Op.36 n°1 — 1° movimento", "Mãos separadas, identificar exposição e desenvolvimento"},
					{15, "pop", "Pirates of the Caribbean — melodia", "Aprender a melodia do tema principal"},
					{10, "teoria", "Forma sonata: exposição/desenvolvimento/recapitulação", ""},
				}},
				{Name: "Terça-feira", Order: 2, Focus: "Lirismo + Cinema", Tasks: []task{
					{15, "tecnica", "Escalas menores BPM 80", ""},
					{20, "tecnica", "Hanon n°21", ""},
					{30, "classico", "Burgmüller n°14 e n°15", ""},
					{30, "pop", "Titanic + Beauty and the Beast — polimento final", "Peças prontas para \"performance\""},
					{20, "anime", "FMA — Bratja (melodia)", "Ouvir, aprender o tema"},
					{10, "teoria", "Sight-reading", ""},
				}},
				{Name: "Quarta-feira", Order: 3, Focus: "Clementi + Técnica", Tasks: []task{
					{15, "tecnica", "Escalas em oitavas — Dó maior, 1 oitava", "Novo desafio! Devagar e sem tensão no pulso"},
					{20, "tecnica", "Hanon n°22", ""},
					{25, "tecnica", "Czerny Op.599 — revisão avançada", ""},
					{35, "classico", "Clementi Sonatina — mãos juntas (16 compassos)", ""},
					{15, "pop", "Pirates — melodia + baixo", ""},
					{10, "teoria", "Cadências: autêntica, plagal, suspensa", ""},
				}},
				{Name: "Quinta-feira", Order: 4, Focus: "Anime Avançado", Tasks: []task{
					{15, "tecnica", "Arpejos + oitavas", ""},
					{20, "tecnica", "Hanon n°23", ""},
					{30, "classico", "Bach Invenção n°1 — polimento expressivo", "Articulação e caráter barroco"},
					{30, "anime", "Attack on Titan — Guren no Yumiya (melodia)", "Ouvir o arranjo, aprender melodia"},
					{20, "rock", "Piano Man — performance", "Peça completa e expressiva"},
					{10, "teoria", "Escalas menores de Si e Fá#", ""},
				}},
				{Name: "Sexta-feira", Order: 5, Focus: "Revisão + Rock", Tasks: []task{
					{15, "tecnica", "Revisão semanal", ""},
					{20, "tecnica", "Hanon n°1–23 seleção", "5 peças favoritas hoje"},
					{30, "classico", "Mini-recital clássico", ""},
					{30, "pop", "Harry Potter, Pirates, Circle of Life", ""},
					{15, "rock", "November Rain — intro (ouvir + tentar primeiros 4 compassos)", ""},
				}},
				{Name: "Sábado", Order: 6, Focus: "Livre + Teoria", Tasks: []task{
					{30, "tecnica", "Oitavas e arpejos (dedicação especial)", ""},
					{30, "classico", "Clementi Sonatina — avançar", ""},
					{30, "anime", "Spirited Away — performance", "Peça pronta para \"conquistado\""},
					{30, "teoria", "Análise harmônica da Sonatina de Clementi", ""},
				}},
			},
		},
		{
			Number: 8,
			Title:  "Fim do Bimestre 2",
			Theme:  "Consolidação, gravação de todas as peças dominadas",
			Days: []day{
				{Name: "Segunda-feira", Order: 1, Focus: "Consolidação Técnica", Tasks: []task{
					{20, "tecnica", "Escalas: todas as 12 maiores (BPM 80)", "Ciclo das quintas completo"},
					{25, "tecnica", "Hanon n°1–10 (BPM 90)", ""},
					{25, "tecnica", "Czerny Op.261 — selecionar 5 difíceis", ""},
					{30, "classico", "Clementi Sonatina — mãos juntas completo", ""},
					{10, "anime", "AoT — avançar melodia", ""},
					{10, "teoria", "Revisão teoria bimestre 2", ""},
				}},
				{Name: "Terça-feira", Order: 2, Focus: "Polimento Pop", Tasks: []task{
					{15, "tecnica", "Arpejos BPM 90", ""},
					{20, "tecnica", "Hanon n°11–20 (BPM 80)", ""},
					{30, "pop", "Passagem todas as peças pop/cinema", "Titanic, Beauty, Harry Potter, Pirates, Circle of Life"},
					{30, "anime", "Passagem todas as peças anime", "Pokémon, Naruto, Dragon Ball, Spirited Away"},
					{15, "rock", "Imagine + Piano Man — performance", ""},
					{10, "teoria", "Sight-reading", ""},
				}},
				{Name: "Quarta-feira", Order: 3, Focus: "Gravação", Tasks: []task{
					{15, "tecnica", "Aquecimento leve", ""},
					{25, "classico", "GRAVAR: Bach Prelúdios n°1 e n°2", "Gravar com celular, ouvir e anotar melhorias"},
					{25, "pop", "GRAVAR: Titanic + Harry Potter", ""},
					{25, "anime", "GRAVAR: Naruto + Spirited Away", ""},
					{20, "rock", "GRAVAR: Imagine + Piano Man", ""},
					{10, "teoria", "Análise das gravações", ""},
				}},
				{Name: "Quinta-feira", Order: 4, Focus: "Pontos Fracos", Tasks: []task{
					{15, "tecnica", "Técnica: exercícios nos pontos fracos", "Escolha 2 dificuldades identificadas nas gravações"},
					{20, "tecnica", "Hanon específico para o problema", ""},
					{30, "classico", "Bach Invenção n°1 — polimento final", ""},
					{30, "pop", "Peças que precisam de mais trabalho", ""},
					{25, "anime", "AoT + FMA Bratja — continuar", ""},
					{10, "teoria", "Teoria: ponto orgânico do mês", ""},
				}},
				{Name: "Sexta-feira", Order: 5, Focus: "Grande Recital Pessoal", Tasks: []task{
					{10, "tecnica", "Aquecimento curto", ""},
					{110, "classico", "GRANDE RECITAL — bimestre 2", "Escalas → Hanon → Bach → Clementi → Burgmüller → Pop → Anime → Rock. GRAVE TUDO!"},
				}},
				{Name: "Sábado", Order: 6, Focus: "Descanso e Teoria", Tasks: []task{
					{30, "tecnica", "Técnica leve: apenas escalas", "Recuperação muscular"},
					{30, "classico", "Leitura musical: explorar peça nova", ""},
					{30, "teoria", "Revisão completa da Fase 1 até aqui", ""},
					{30, "anime", "Assistir performance de piano no YouTube", "Joe Hisaishi, Lang Lang — inspiração!"},
				}},
			},
		},
		{
			Number: 9,
			Title:  "Introdução à Escola de Velocidade",
			Theme:  "Czerny Op.849, escala em terças, La La Land",
			Days: []day{
				{Name: "Segunda-feira", Order: 1, Focus: "Op.849 + Terças", Tasks: []task{
					{15, "tecnica", "Escalas em terças — Dó maior", "Novo! MDir: terças paralelas, BPM 50"},
					{25, "tecnica", "Czerny Op.849 n°1 (novo)", "Escola de Velocidade — mãos separadas"},
					{25, "tecnica", "Hanon n°24–25", ""},
					{30, "classico", "Bach Invenção n°4 BWV 775 (leitura)", "Nova invenção! MDir isolada"},
					{15, "pop", "La La Land — City of Stars (melodia)", ""},
					{10, "teoria", "Intervalos: terça, quarta, quinta, sexta", ""},
				}},
				{Name: "Terça-feira", Order: 2, Focus: "Pop + Expressão", Tasks: []task{
					{15, "tecnica", "Arpejos menores (Lá, Mi, Ré)", ""},
					{20, "tecnica", "Hanon n°26", ""},
					{30, "classico", "Clementi Sonatina — polimento expressivo", "Dinâmicas, pedalização básica"},
					{30, "pop", "City of Stars — melodia + baixo", ""},
					{20, "anime", "AoT + FMA — polimento", ""},
					{10, "teoria", "Sight-reading", ""},
				}},
				{Name: "Quarta-feira", Order: 3, Focus: "Op.849 + Invenção", Tasks: []task{
					{15, "tecnica", "Escalas em terças BPM 60", ""},
					{20, "tecnica", "Czerny Op.849 n°1 — mãos juntas", ""},
					{25, "tecnica", "Hanon n°27", ""},
					{35, "classico", "Bach Invenção n°4 — MEsq isolada", "Aprender a voz inferior completa"},
					{15, "pop", "City of Stars — mãos juntas BPM 50", ""},
					{10, "teoria", "Inversão de acordes", ""},
				}},
				{Name: "Quinta-feira", Order: 4, Focus: "Anime + Rock", Tasks: []task{
					{15, "tecnica", "Revisão técnica", ""},
					{20, "tecnica", "Hanon n°28", ""},
					{30, "classico", "Burgmüller n°16 e n°17", ""},
					{30, "anime", "Howl's Moving Castle — Merry-Go-Round (melodia)", "Peça mágica de Joe Hisaishi"},
					{20, "rock", "Clocks — polimento", ""},
					{10, "teoria", "Escala menor harmônica de Si", ""},
				}},
				{Name: "Sexta-feira", Order: 5, Focus: "Revisão", Tasks: []task{
					{15, "tecnica", "Aquecimento", ""},
					{20, "tecnica", "Hanon + Czerny Op.849 revisão", ""},
					{30, "classico", "Mini-recital", ""},
					{30, "pop", "City of Stars + La La Land performance", ""},
					{15, "teoria", "Auto-avaliação", ""},
				}},
				{Name: "Sábado", Order: 6, Focus: "Teoria + Aprofundamento", Tasks: []task{
					{25, "tecnica", "Terças paralelas + arpejos", ""},
					{30, "classico", "Bach Invenção n°4 — mãos juntas tentativa", ""},
					{35, "pop", "Howl's Moving Castle — avançar", ""},
					{30, "teoria", "Análise do Bach Invenção n°4", "Imitação, sequências, cadências"},
				}},
			},
		},
		{
			Number: 10,
			Title:  "Dois Caminhos: Clássico e Moderno",
			Theme:  "Bach Invenção n°4 mãos juntas, Interstellar, November Rain intro",
			Days: []day{
				{Name: "Segunda-feira", Order: 1, Focus: "Invenção + Velocidade", Tasks: []task{
					{15, "tecnica", "Escalas menores BPM 90", ""},
					{25, "tecnica", "Czerny Op.849 n°2 (novo)", ""},
					{25, "tecnica", "Hanon n°29–30", ""},
					{35, "classico", "Bach Invenção n°4 — mãos juntas BPM 55", ""},
					{10, "pop", "Interstellar — Main Theme (exploração)", "Aprender os primeiros 8 compassos"},
					{10, "teoria", "Passagem de polegar nas escalas", ""},
				}},
				{Name: "Terça-feira", Order: 2, Focus: "Cinema + Emoção", Tasks: []task{
					{15, "tecnica", "Arpejos menores BPM 90", ""},
					{20, "tecnica", "Hanon n°21–25 revisão", ""},
					{30, "classico", "Clementi Sonatina — polimento final", ""},
					{30, "pop", "Interstellar — avançar melodia", ""},
					{20, "anime", "Merry-Go-Round — mãos juntas BPM 50", ""},
					{10, "teoria", "Sight-reading", ""},
				}},
				{Name: "Quarta-feira", Order: 3, Focus: "Técnica Avançada", Tasks: []task{
					{15, "tecnica", "Escalas em 6as paralelas — Dó", "Mais difícil que terças, BPM 50"},
					{20, "tecnica", "Czerny Op.849 n°2 — mãos juntas", ""},
					{25, "tecnica", "Hanon n°31 (novo)", ""},
					{35, "classico", "Bach Invenção n°4 — BPM 65, expressão", ""},
					{15, "rock", "November Rain — intro (16 compassos)", "Os primeiros 16 compassos do piano solo"},
					{10, "teoria", "Modulação simples", ""},
				}},
				{Name: "Quinta-feira", Order: 4, Focus: "Anime Completo", Tasks: []task{
					{15, "tecnica", "Revisão técnica", ""},
					{20, "tecnica", "Hanon n°32", ""},
					{30, "classico", "Bach Prelúdio n°3 BWV 848 (leitura)", "Novo prelúdio. Apenas MDir hoje"},
					{30, "anime", "Guren no Yumiya (AoT) — mãos juntas", ""},
					{20, "pop", "City of Stars — polimento final", ""},
					{10, "teoria", "Acorde diminuto", ""},
				}},
				{Name: "Sexta-feira", Order: 5, Focus: "Revisão + Rock", Tasks: []task{
					{15, "tecnica", "Revisão semanal", ""},
					{20, "tecnica", "Op.849 n°1 e n°2 revisão", ""},
					{30, "classico", "Mini-recital", ""},
					{30, "pop", "Interstellar, City of Stars, Pirates", ""},
					{15, "rock", "November Rain intro + Clocks: ambas", ""},
				}},
				{Name: "Sábado", Order: 6, Focus: "Exploração + Teoria", Tasks: []task{
					{30, "tecnica", "Trinados: exercício (Dó-Ré, BPM 60)", ""},
					{30, "classico", "Bach Invenções n°1 e n°4 — performance", "Duas invenções seguidas"},
					{30, "rock", "November Rain — avançar", ""},
					{30, "teoria", "Ornamentos barrocos: mordente, grupeto", ""},
				}},
			},
		},
		{
			Number: 11,
			Title:  "Quase Lá — Pré-Fase 2",
			Theme:  "Czerny Op.849 n°3, Bach Prelúdio n°3, Your Lie in April",
			Days: []day{
				{Name: "Segunda-feira", Order: 1, Focus: "Op.849 + Prelúdio n°3", Tasks: []task{
					{15, "tecnica", "Escalas maiores e menores alternadas", "Uma maior, uma menor, BPM 90"},
					{25, "tecnica", "Czerny Op.849 n°3 (novo)", ""},
					{25, "tecnica", "Hanon n°33–34", ""},
					{35, "classico", "Bach Prelúdio n°3 — mãos juntas BPM 55", ""},
					{10, "anime", "Your Lie in April — Hikaru Nara (melodia)", "Ouvir, aprender o tema principal"},
					{10, "teoria", "Modo jônico e eólio (maior/menor)", ""},
				}},
				{Name: "Terça-feira", Order: 2, Focus: "Anime + Cinema", Tasks: []task{
					{15, "tecnica", "Arpejos completos BPM 90", ""},
					{20, "tecnica", "Hanon n°35", ""},
					{30, "classico", "Burgmüller n°18 e n°19", ""},
					{30, "anime", "Hikaru Nara — avançar", ""},
					{20, "pop", "Interstellar — polimento", ""},
					{10, "teoria", "Sight-reading", ""},
				}},
				{Name: "Quarta-feira", Order: 3, Focus: "Consolidação Técnica", Tasks: []task{
					{15, "tecnica", "Trinados: 5 minutos", ""},
					{20, "tecnica", "Op.849 n°1–3 revisão", ""},
					{25, "tecnica", "Hanon n°36", ""},
					{35, "classico", "Bach Invenção n°4 — BPM 70, expressão barroca", ""},
					{15, "pop", "A Thousand Years — Christina Perri (melodia)", ""},
					{10, "teoria", "Análise do Prelúdio n°3 de Bach", ""},
				}},
				{Name: "Quinta-feira", Order: 4, Focus: "Rock Avançado", Tasks: []task{
					{15, "tecnica", "Revisão técnica", ""},
					{20, "tecnica", "Hanon n°37–38", ""},
					{30, "classico", "Clementi Sonatina Op.36 n°1 — PERFORMANCE", "Do início ao fim com confiança"},
					{30, "rock", "November Rain — além da intro", ""},
					{20, "anime", "Merry-Go-Round — polimento", ""},
					{10, "teoria", "Progressão I-V-vi-IV (pop)", "A mais comum do pop moderno"},
				}},
				{Name: "Sexta-feira", Order: 5, Focus: "Revisão Geral", Tasks: []task{
					{15, "tecnica", "Aquecimento completo", ""},
					{20, "tecnica", "Hanon + Czerny seleção", ""},
					{30, "classico", "Mini-recital: Bach + Clementi", ""},
					{30, "pop", "Recital pop/anime/rock completo", ""},
					{15, "teoria", "Auto-avaliação final da semana", ""},
				}},
				{Name: "Sábado", Order: 6, Focus: "Preparação Fase 2", Tasks: []task{
					{20, "tecnica", "TODAS as 12 escalas maiores + 4 menores (BPM 90)", "Conclusão do trabalho de escalas da Fase 1"},
					{30, "classico", "Bach: Prelúdios + Invenções — performance completa", ""},
					{30, "pop", "Peça favorita da Fase 1 — performance dedicada", ""},
					{20, "teoria", "Planejar: peças da Fase 2", "Rever o roadmap, definir próximas obras"},
				}},
			},
		},
		{
			Number: 12,
			Title:  "Grande Finale — Fase 1 Completa!",
			Theme:  "Recital completo, polimento final, celebração",
			Days: []day{
				{Name: "Segunda-feira", Order: 1, Focus: "Polimento Final", Tasks: []task{
					{20, "tecnica", "Escalas: 12 maiores (BPM 90–100)", "Meta final da Fase 1"},
					{25, "tecnica", "Hanon n°1–10 (BPM 90)", ""},
					{25, "tecnica", "Czerny Op.849 n°1–3 — polimento", ""},
					{30, "classico", "Bach: Prelúdios n°1–3 + Invenções n°1 e n°4", "Polimento total"},
					{10, "anime", "Hikaru Nara — mãos juntas", ""},
					{10, "teoria", "Revisão de teoria da Fase 1", ""},
				}},
				{Name: "Terça-feira", Order: 2, Focus: "Cinema e Pop Final", Tasks: []task{
					{15, "tecnica", "Aquecimento leve", ""},
					{20, "tecnica", "Arpejos + terças — polimento", ""},
					{35, "pop", "POLIMENTO: Titanic, Beauty, Harry Potter, Pirates", "Expressão, dinâmica e musicalidade"},
					{30, "pop", "POLIMENTO: Circle of Life, Interstellar, City of Stars", ""},
					{20, "teoria", "Revisão de harmonia", ""},
				}},
				{Name: "Quarta-feira", Order: 3, Focus: "Anime Final", Tasks: []task{
					{15, "tecnica", "Aquecimento", ""},
					{20, "tecnica", "Ponto fraco da semana", ""},
					{35, "anime", "POLIMENTO: Pokémon, Naruto, Dragon Ball, Spirited Away", ""},
					{30, "anime", "POLIMENTO: Merry-Go-Round, AoT, FMA, Hikaru Nara", ""},
					{20, "teoria", "Análise harmônica de uma peça pop", ""},
				}},
				{Name: "Quinta-feira", Order: 4, Focus: "Rock Final", Tasks: []task{
					{15, "tecnica", "Aquecimento", ""},
					{20, "tecnica", "Hanon n°1–5 (BPM 100)", "Máxima velocidade da Fase 1"},
					{35, "rock", "POLIMENTO: Imagine, Piano Man, Clocks", ""},
					{30, "rock", "POLIMENTO: November Rain, Let It Be, Yesterday", ""},
					{20, "teoria", "Objetivos da Fase 2", ""},
				}},
				{Name: "Sexta-feira", Order: 5, Focus: "GRANDE RECITAL", Tasks: []task{
					{10, "tecnica", "Aquecimento mínimo", "5 escalas e 2 Hanon — não se canse"},
					{110, "classico", "GRANDE RECITAL DA FASE 1", "Escalas → Hanon → Bach Prelúdios → Bach Invenções → Clementi → Burgmüller → Titanic → Beauty → Harry Potter → Pirates → Circle of Life → Interstellar → Pokémon → Naruto → Dragon Ball → Spirited Away → Merry-Go-Round → AoT → Imagine → Piano Man → Clocks → November Rain. GRAVE TUDO!"},
				}},
				{Name: "Sábado", Order: 6, Focus: "Celebração + Fase 2", Tasks: []task{
					{30, "tecnica", "Escalas livres — toque o que quiser", "Você completou a Fase 1. Parabéns!"},
					{30, "classico", "Leitura exploratória: Haydn Sonata, Mozart K.545", "Primeiros passos para a Fase 2"},
					{30, "pop", "ESCOLHA LIVRE: toque sua peça favorita", "Você merece!"},
					{30, "teoria", "Planejamento: definir peças da Fase 2", "Consulte o roadmap e marque os objetivos"},
				}},
			},
		},
	}
}

func pieces() []piece {
	return []piece{
		// TÉCNICA
		{"Hanon n°1-10", "Hanon", "tecnica", 2, 1, 100, "", 1},
		{"Czerny Op.599 n°1-10", "Czerny", "tecnica", 2, 1, 100, "", 2},
		{"Czerny Op.261 n°1-10", "Czerny", "tecnica", 3, 1, 150, "", 3},
		{"Hanon n°11-30", "Hanon", "tecnica", 4, 1, 200, "", 4},
		{"Czerny Op.261 n°11-20", "Czerny", "tecnica", 4, 1, 200, "", 5},
		{"Czerny Op.849 n°1-5", "Czerny", "tecnica", 5, 2, 300, "", 6},
		{"Hanon n°31-60", "Hanon", "tecnica", 5, 2, 300, "", 7},
		{"Czerny Op.849 n°6-15", "Czerny", "tecnica", 6, 3, 400, "", 8},
		{"Czerny Op.740 n°1-10", "Czerny", "tecnica", 7, 3, 500, "", 9},
		// CLÁSSICO/BARROCO
		{"Burgmüller Op.100 n°1-5", "Burgmüller", "classico", 2, 1, 150, "https://imslp.org/wiki/25_Progressive_Studies,_Op.100_(Burgm%C3%BCller,_Johann_Friedrich)", 10},
		{"Bach Prelúdio n°1 BWV 846", "J.S. Bach", "classico", 3, 1, 200, "https://imslp.org/wiki/Prelude_and_Fugue_in_C_major,_BWV_846_(Bach,_Johann_Sebastian)", 11},
		{"Burgmüller Op.100 n°6-15", "Burgmüller", "classico", 3, 1, 200, "", 12},
		{"Bach Prelúdio n°2 BWV 847", "J.S. Bach", "classico", 4, 1, 250, "", 13},
		{"Clementi Sonatina Op.36 n°1", "Clementi", "classico", 4, 1, 300, "https://imslp.org/wiki/6_Sonatinas,_Op.36_(Clementi,_Muzio)", 14},
		{"Bach Invenção n°1 BWV 772", "J.S. Bach", "classico", 4, 2, 300, "https://imslp.org/wiki/Inventions_and_Sinfonias,_BWV_772-801_(Bach,_Johann_Sebastian)", 15},
		{"Bach Invenção n°4 BWV 775", "J.S. Bach", "classico", 5, 2, 350, "", 16},
		{"Haydn Sonata Hob.XVI/1", "Haydn", "classico", 4, 2, 350, "", 17},
		{"Scarlatti Sonata K.1", "Scarlatti", "classico", 4, 2, 300, "", 18},
		{"Mozart Sonata K.545", "Mozart", "classico", 5, 3, 500, "https://imslp.org/wiki/Piano_Sonata_No.16_in_C_major,_K.545_(Mozart,_Wolfgang_Amadeus)", 19},
		{"Beethoven Sonata Op.49 n°1", "Beethoven", "classico", 5, 3, 500, "", 20},
		{"Beethoven Für Elise", "Beethoven", "classico", 5, 3, 400, "", 21},
		{"Chopin Prelúdio Op.28 n°7", "Chopin", "classico", 3, 3, 300, "", 22},
		{"Chopin Prelúdio Op.28 n°4", "Chopin", "classico", 5, 4, 500, "", 23},
		{"Beethoven Moonlight Sonata 1° mv", "Beethoven", "classico", 6, 4, 700, "", 24},
		{"Chopin Noturno Op.9 n°2", "Chopin", "classico", 6, 4, 700, "", 25},
		// POP/CINEMA
		{"Imagine", "John Lennon", "pop", 3, 1, 200, "", 26},
		{"Let It Be", "The Beatles", "pop", 3, 1, 200, "", 27},
		{"River Flows in You", "Yiruma", "pop", 4, 1, 200, "", 28},
		{"Titanic – My Heart Will Go On", "J. Horner/arr.", "pop", 4, 1, 250, "", 29},
		{"Beauty and the Beast", "A. Menken/arr.", "pop", 4, 1, 200, "", 30},
		{"Harry Potter – Hedwig's Theme", "J. Williams/arr.", "pop", 4, 1, 200, "", 31},
		{"Piano Man", "Billy Joel", "pop", 4, 1, 250, "", 32},
		{"Pokémon Main Theme", "arr.", "anime", 3, 1, 150, "", 33},
		{"Naruto – Sadness and Sorrow", "T. Masuda/arr.", "anime", 3, 1, 200, "", 34},
		{"Spirited Away – One Summer's Day", "Joe Hisaishi", "anime", 4, 1, 300, "", 35},
		{"Circle of Life", "E. John/arr.", "pop", 4, 2, 250, "", 36},
		{"He's a Pirate (Piratas)", "H. Zimmer/arr.", "pop", 5, 2, 300, "", 37},
		{"Clocks", "Coldplay", "pop", 4, 2, 250, "", 38},
		{"A Thousand Years", "C. Perri/arr.", "pop", 4, 2, 250, "", 39},
		{"City of Stars (La La Land)", "Hurwitz/arr.", "pop", 5, 2, 300, "", 40},
		{"Interstellar Main Theme", "H. Zimmer/arr.", "pop", 5, 2, 300, "", 41},
		{"Dragon Ball Z – Cha-La", "arr.", "anime", 4, 2, 250, "", 42},
		{"Howl's Moving Castle", "Joe Hisaishi", "anime", 5, 2, 350, "", 43},
		{"Attack on Titan – Guren no Yumiya", "arr.", "anime", 5, 2, 300, "", 44},
		{"FMA – Bratja", "arr.", "anime", 4, 2, 250, "", 45},
		{"Your Lie in April – Hikaru Nara", "Goose House/arr.", "anime", 5, 2, 350, "", 46},
		{"November Rain (intro)", "Guns N' Roses/arr.", "pop", 5, 2, 400, "", 47},
		{"Bohemian Rhapsody (simplificado)", "Queen/arr.", "pop", 5, 3, 500, "", 48},
		{"Comptine d'un autre été (Amélie)", "Y. Tiersen", "pop", 4, 3, 250, "", 49},
		{"Nuvole Bianche", "Einaudi", "pop", 5, 4, 400, "", 50},
	}
}
