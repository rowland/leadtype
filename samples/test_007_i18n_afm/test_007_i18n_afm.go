// Copyright 2014 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package main

import (
	"os"
	"os/exec"

	"github.com/rowland/leadtype/afm_fonts"
	"github.com/rowland/leadtype/options"
	"github.com/rowland/leadtype/pdf"
)

import "fmt"

const name = "test_007_i18n_afm.pdf"

func main() {
	doc := pdf.NewDocWriter()
	afmfc, err := afm_fonts.New("../../afm/data/fonts/*.afm")
	if err != nil {
		panic(err)
	}
	doc.AddFontSource(afmfc)
	doc.SetUnits("in")

	doc.NewPage()
	doc.MoveTo(1, 1)
	_, err = doc.SetFont("Helvetica", 12, options.Options{})
	if err != nil {
		panic(err)
	}
	doc.SetUnderline(true)
	doc.Print("I18N Text\n\n")
	doc.SetUnderline(false)
	doc.SetLineSpacing(1.2)

	for _, k := range i18nKeys {
		// fmt.Println(i18nText[k])
		if doc.Y() > 10 {
			doc.NewPage()
			doc.MoveTo(1, 1)
		}
		fmt.Fprintf(doc, "%s:\n", k)
		doc.PrintWithOptions(i18nText[k], options.Options{"width": 6.5})
		fmt.Fprintln(doc)
	}
	f, err := os.Create(name)
	if err != nil {
		panic(err)
	}
	doc.WriteTo(f)
	f.Close()
	exec.Command("open", name).Start()
}

var i18nKeys = []string{"cs", "da", "de", "en", "es", "es_US", "fi", "fr", "hr", "hu", "id", "it", "nl", "no", "pt", "pt_BR", "sl", "sv", "tr"}
var i18nText = map[string]string{
	// "ar":    "أثبتت الدراسات التي اجريت على الأنماط السلوكية أن الأشخاص الأكثر فاعلية هم الذين يفهمون أنفسهم ونقاط القوى والضعف لديهم, بالتالي تصبح لديهم القدرة على ابتكار أساليب واستراتيجيات معينة لتحقيق مطالب البيئة المحيطة بهم.",
	"cs": "Výsledky výzkumu chování ukazují, že nejefektivnějšími lidmi jsou ti, kdo znají sebe sama a uvědomují si své silné i slabé stránky. Jsou schopni vyvinout strategie, jejichž pomocí se vyrovnávají s požadavky, které na ně klade okolí.",
	"da": "Adfærdsforskningen mener, at de mest effektive mennesker er de, der forstår sig selv, dvs. mennesker, der forstår såvel egne styrker som svagheder og på baggrund heraf kan udvikle egne fremgangsmåder til at tilfredsstille omgivelsernes krav.",
	"de": "Die Verhaltensforschung ist der Ansicht, dass die effektivsten Menschen jene sind, die sich selbst kennen, sowohl ihre Stärken als auch ihre Schwächen, so dass sie Strategien entwickeln können, um den Anforderungen ihres Umfeldes gerecht zu werden.",
	// "el":    "Από έρευνες συμπεριφοράς προκύπτει ότι πιό αποτελεσματικά είναι τα άτομα που έχουν επίγνωση του εαυτού τους, που γνωρίζουν τόσο τα δυνατά τους σημεία, όσο και τις αδυναμίες τους και μπορούν να αναπτύξουν στρατηγικές που καλύπτουν τις ανάγκες του περιβάλλοντός τους.",
	"en":    "Behavioral research suggests that the most effective people are those who understand themselves, both their strengths and weaknesses, so they can develop strategies to meet the demands of their environment.",
	"es":    "La investigación sobre el comportamiento sugiere que las personas más efectivas son las que se conocen a sí mismas, sus fortalezas y debilidades, por lo que son capaces de desarrollar estrategias que den respuesta a las demandas de su entorno.",
	"es_US": "La investigación sobre el comportamiento sugiere que las personas más efectivas son aquellas que se conocen integralmente a sí mismos, saben cuáles son sus habilidades y debilidades, y así tienen la posibilidad de desarrollar estrategias que cumplan las demandas de su entorno.",
	// "fa":    "پژوهش‌هاي رفتاری نشان می‌دهند که مؤثرترين افراد آنهايي هستند که خودشان را می‌شناسند و با نقاط قوت و ضعف خود آشنايي دارند، لذا می‌توانند استراتژی‌هايي مطابق با نياز محيط خود تدوين نمايند.",
	"fi": "Käyttäytymistutkimuksen mukaan tehokkaimpia ihmisiä ovat ne, jotka tunnistavat sekä vahvat että heikot puolensa. Heillä on taito kehittää työympäristön vaatimat strategiat.",
	"fr": "La recherche sur le comportement humain indique que les personnes qui réussissent le mieux, sont celles qui se connaissent, qui connaissent leurs atouts et leurs faiblesses, de sorte qu'elles peuvent mettre au point des stratégies pour répondre aux exigences de leur environnement.",
	"hr": "Rezultati na temelju znanstvenih istraživanja ljudskog ponašanja govore da su najefikasniji oni ljudi, koji poznaju sebe, kako svoje prednosti tako i svoje nedostatke, te su u stanju na osnovu tih spoznaja razvijati strategije da bi mogli zadovoljiti očekivanja okoline.",
	"hu": "A viselkedéstudomány szerint a legeredményesebb emberek azok, akik tisztában vannak saját magukkal, ismerik erősségeiket és gyengeségeiket, így stratégiákat tudnak kidolgozni arra, hogyan feleljenek meg a környezetük támasztotta követelményeknek.",
	"id": "Penelitian perilaku menunjukkan bahwa orang yang paling efektif adalah mereka yang memahami diri sendiri, baik kelebihan dan kekurangan mereka, sehingga mereka dapat mengembangkan strategi untuk memenuhi tuntutan lingkungan mereka.",
	"it": "Secondo la ricerca comportamentale, le persone più efficienti sono quelle che comprendono se stesse, che conoscono, cioè, i propri punti di forza e le aree di miglioramento e che sono in grado di sviluppare le strategie più idonee per far fronte alle esigenze dettate dall'ambiente che le circonda.",
	// "ja":    "行動学の研究によると、最も有能な人というのは、自分のことがよくわかっている人、つまり自分の長所も弱点も知りぬいており、自分の置かれた環境が要求する物事に対応できるような戦略を構築できる人のことだといわれています。",
	"nl": "Uitkomsten uit gedragsonderzoek tonen aan dat de effectiviteit van mensen toeneemt naarmate zij zichzelf beter kennen en begrijpen. Het herkennen van sterkten en zwakten biedt de kans strategieën en manieren te ontwikkelen die aan de eisen van de omgeving voldoen.",
	"no": "Adferdsforskning hevder at de mest effektive menneskene er de som har forståelse for hvordan de selv fungerer og handler i ulike situasjoner. De kjenner sine sterke og svake sider, og kan derfor utvikle strategier for å møte de krav omgivelsene stiller til dem.",
	// missing "ń"
	// "pl":     "Badania zachowań ludzkich sugerują wniosek, że najskuteczniejsi w działaniu są ci ludzie, którzy rozumieją samych siebie, znają zarówno swoje mocne jak i słabe strony - a więc potrafią opracowywać strategie działania, spełniające oczekiwania ich otoczenia.",
	"pt":    "Os estudos realizados sobre o comportamento sugerem que as pessoas mais eficazes são aquelas que se conhecem a si próprias, tanto no que respeita aos seus pontos fortes, como aos seus pontos fracos. Por isso, estão mais aptas a desenvolver estratégias adequadas às exigências do seu meio envolvente.",
	"pt_BR": "Estudos realizados sobre o comportamento sugerem que os indivíduos mais eficazes são aqueles que conhecem a si próprios, tanto nos seus pontos fortes quanto nos fracos. Por isso, estão mais aptos a desenvolver estratégias adequadas às exigências de seu meio ambiente.",
	// "ru":     "Исследования в области поведения указывают на то, что наибольшего успеха добиваются люди, которые знают самих себя, как свои сильные, так и слабые стороны. На основе этого они смогут разработать собственную стратегию поведения, позволяющую наилучшим образом соответствовать тем требованиям, которые к ним предъявляет среда.",
	"sl": "Vedenjske raziskave kažejo, da najbolj učinkoviti ljudje poznajo sami sebe, svoje prednosti in slabosti. Le na ta način lahko razvijejo sposobnosti, da bodo zadovoljili zahteve svojega okolja. To poročilo analizira vaš vedenjski stil; na kakšen način in kako delate določene stvari, ter opredeljuje samo vaše vedenje.",
	"sv": "Beteendeforskningen hävdar att de effektivaste människorna är de som har förståelse för hur de själva fungerar och agerar i olika situationer. De känner sina starka och svaga sidor och kan därför utveckla egna tillvägagångssätt för att möta omgivningens krav.",
	// "th":     "การวิจัยด้านพฤติกรรมศาสตร์ชี้แนะว่า ผู้ที่มีประสิทธิผลสูงสุดคือผู้ที่เข้าใจในตนเอง ทั้งในเรื่องจุดแข็งและจุดอ่อนของตน ด้วยความเข้าใจนี้ เขาสามารถพัฒนากลยุทธ์เพื่อสนองต่ออุปสงค์หรือความต้องการของสิ่งแวดล้อมของเขาได้",
	"tr": "Davranış üzerine yapılan araştırmalar, insanların en etkin olanlarının kendilerini anlayanlar, hem kendi güçlü yönlerini ve hem de zaaflarını kavrayarak çevrelerinin taleplerini karşılayan stratejileri geliştirebilen kişiler olduklarını göstermektedir.",
	// "vi":     "Nghiên cứu hành vi cho thấy người hiệu quả nhất chính là người hiểu rõ bản thân họ nhất - cả điểm yếu lẫn điểm mạnh, từ đó họ có thể phát triển các chiến lược để đáp ứng yêu cầu của môi trường.",
	// "zh-chs": "行为研究表明：效率最高的人是那些了解自己长处和短处的人，从而他们可以制定战略，适应环境要求。",
	// "zh-cht": "針對行為的研究發現，越了解自己優缺點的人，越能夠制訂適當策略，配合環境所需，也就越有可能成功。",
}
