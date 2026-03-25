// Copyright 2014 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package main

import (
	"fmt"

	"github.com/rowland/leadtype/afm_fonts"
	"github.com/rowland/leadtype/options"
	"github.com/rowland/leadtype/pdf"
)

func init() {
	registerSample("test_007_i18n_afm", "render international AFM text samples", runTest007I18nAfm, "-a", "Firefox")
}

func runTest007I18nAfm() (string, error) {
	return writeDoc("test_007_i18n_afm.pdf", func(doc *pdf.DocWriter) error {
		afmfc, err := afm_fonts.Default()
		if err != nil {
			return err
		}
		doc.AddFontSource(afmfc)
		doc.SetUnits("in")

		doc.NewPage()
		doc.MoveTo(1, 1)
		if _, err := doc.SetFont("Helvetica", 12, options.Options{}); err != nil {
			return err
		}
		doc.SetUnderline(true)
		doc.Print("I18N Text\n\n")
		doc.SetUnderline(false)
		doc.SetLineSpacing(1.2)

		for _, k := range i18nKeys {
			if doc.Y() > 10 {
				doc.NewPage()
				doc.MoveTo(1, 1)
			}
			fmt.Fprintf(doc, "%s:\n", k)
			doc.PrintWithOptions(i18nText[k], options.Options{"width": 6.5})
			fmt.Fprintln(doc)
		}
		return nil
	})
}

var i18nKeys = []string{"cs", "da", "de", "en", "es", "es_US", "fi", "fr", "hr", "hu", "id", "it", "nl", "no", "pt", "pt_BR", "sl", "sv", "tr"}

var i18nText = map[string]string{
	"cs":    "Výsledky výzkumu chování ukazují, že nejefektivnějšími lidmi jsou ti, kdo znají sebe sama a uvědomují si své silné i slabé stránky. Jsou schopni vyvinout strategie, jejichž pomocí se vyrovnávají s požadavky, které na ně klade okolí.",
	"da":    "Adfærdsforskningen mener, at de mest effektive mennesker er de, der forstår sig selv, dvs. mennesker, der forstår såvel egne styrker som svagheder og på baggrund heraf kan udvikle egne fremgangsmåder til at tilfredsstille omgivelsernes krav.",
	"de":    "Die Verhaltensforschung ist der Ansicht, dass die effektivsten Menschen jene sind, die sich selbst kennen, sowohl ihre Stärken als auch ihre Schwächen, so dass sie Strategien entwickeln können, um den Anforderungen ihres Umfeldes gerecht zu werden.",
	"en":    "Behavioral research suggests that the most effective people are those who understand themselves, both their strengths and weaknesses, so they can develop strategies to meet the demands of their environment.",
	"es":    "La investigación sobre el comportamiento sugiere que las personas más efectivas son las que se conocen a sí mismas, sus fortalezas y debilidades, por lo que son capaces de desarrollar estrategias que den respuesta a las demandas de su entorno.",
	"es_US": "La investigación sobre el comportamiento sugiere que las personas más efectivas son aquellas que se conocen integralmente a sí mismos, saben cuáles son sus habilidades y debilidades, y así tienen la posibilidad de desarrollar estrategias que cumplan las demandas de su entorno.",
	"fi":    "Käyttäytymistutkimuksen mukaan tehokkaimpia ihmisiä ovat ne, jotka tunnistavat sekä vahvat että heikot puolensa. Heillä on taito kehittää työympäristön vaatimat strategiat.",
	"fr":    "La recherche sur le comportement humain indique que les personnes qui réussissent le mieux, sont celles qui se connaissent, qui connaissent leurs atouts et leurs faiblesses, de sorte qu'elles peuvent mettre au point des stratégies pour répondre aux exigences de leur environnement.",
	"hr":    "Rezultati na temelju znanstvenih istraživanja ljudskog ponašanja govore da su najefikasniji oni ljudi, koji poznaju sebe, kako svoje prednosti tako i svoje nedostatke, te su u stanju na osnovu tih spoznaja razvijati strategije da bi mogli zadovoljiti očekivanja okoline.",
	"hu":    "A viselkedéstudomány szerint a legeredményesebb emberek azok, akik tisztában vannak saját magukkal, ismerik erősségeiket és gyengeségeiket, így stratégiákat tudnak kidolgozni arra, hogyan feleljenek meg a környezetük támasztotta követelményeknek.",
	"id":    "Penelitian perilaku menunjukkan bahwa orang yang paling efektif adalah mereka yang memahami diri sendiri, baik kelebihan dan kekurangan mereka, sehingga mereka dapat mengembangkan strategi untuk memenuhi tuntutan lingkungan mereka.",
	"it":    "Secondo la ricerca comportamentale, le persone più efficienti sono quelle che comprendono se stesse, che conoscono, cioè, i propri punti di forza e le aree di miglioramento e che sono in grado di sviluppare le strategie più idonee per far fronte alle esigenze dettate dall'ambiente che le circonda.",
	"nl":    "Uitkomsten uit gedragsonderzoek tonen aan dat de effectiviteit van mensen toeneemt naarmate zij zichzelf beter kennen en begrijpen. Het herkennen van sterkten en zwakten biedt de kans strategieën en manieren te ontwikkelen die aan de eisen van de omgeving voldoen.",
	"no":    "Adferdsforskning hevder at de mest effektive menneskene er de som har forståelse for hvordan de selv fungerer og handler i ulike situasjoner. De kjenner sine sterke og svake sider, og kan derfor utvikle strategier for å møte de krav omgivelsene stiller til dem.",
	"pt":    "Os estudos realizados sobre o comportamento sugerem que as pessoas mais eficazes são aquelas que se conhecem a si próprias, tanto no que respeita aos seus pontos fortes, como aos seus pontos fracos. Por isso, estão mais aptas a desenvolver estratégias adequadas às exigências do seu meio envolvente.",
	"pt_BR": "Estudos realizados sobre o comportamento sugerem que os indivíduos mais eficazes são aqueles que conhecem a si próprios, tanto nos seus pontos fortes quanto nos fracos. Por isso, estão mais aptos a desenvolver estratégias adequadas às exigências de seu meio ambiente.",
	"sl":    "Vedenjske raziskave kažejo, da najbolj učinkoviti ljudje poznajo sami sebe, svoje prednosti in slabosti. Le na ta način lahko razvijejo sposobnosti, da bodo zadovoljili zahteve svojega okolja. To poročilo analizira vaš vedenjski stil; na kakšen način in kako delate določene stvari, ter opredeljuje samo vaše vedenje.",
	"sv":    "Beteendeforskningen hävdar att de effektivaste människorna är de som har förståelse för hur de själva fungerar och agerar i olika situationer. De känner sina starka och svaga sidor och kan därför utveckla egna tillvägagångssätt för att möta omgivningens krav.",
	"tr":    "Davranış üzerine yapılan araştırmalar, insanların en etkin olanlarının kendilerini anlayanlar, hem kendi güçlü yönlerini ve hem de zaaflarını kavrayarak çevrelerinin taleplerini karşılayan stratejileri geliştirebilen kişiler olduklarını göstermektedir.",
}
