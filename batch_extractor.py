import os
import csv
from extractor import extrair_pdf

PDF_DIR = "data/pdfs"
CSV_PATH = "data/csv/output.csv"


def main():
    os.makedirs("data/csv", exist_ok=True)

    arquivos = [f for f in os.listdir(PDF_DIR) if f.endswith(".pdf")]

    resultados = []

    for arquivo in arquivos:
        caminho = os.path.join(PDF_DIR, arquivo)
        print(f"Processando: {arquivo}")

        resultado = extrair_pdf(caminho)
        resultado["arquivo"] = arquivo

        resultados.append(resultado)

    if not resultados:
        print("Nenhum PDF encontrado.")
        return

    campos = sorted({k for r in resultados for k in r.keys()})

    with open(CSV_PATH, "w", newline="", encoding="utf-8") as f:
        writer = csv.DictWriter(f, fieldnames=campos)
        writer.writeheader()
        writer.writerows(resultados)

    print(f"\nCSV salvo em: {CSV_PATH}")


if __name__ == "__main__":
    main()
