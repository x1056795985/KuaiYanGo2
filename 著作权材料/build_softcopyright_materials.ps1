$ErrorActionPreference = 'Stop'

$root = Split-Path -Parent $PSScriptRoot
$outDir = $PSScriptRoot

$sourceSelections = @(
    @{ Path = 'main.go'; Start = 1; End = 87 },
    @{ Path = 'core/core.go'; Start = 1; End = 155 },
    @{ Path = 'core/InitRouter.go'; Start = 1; End = 150 },
    @{ Path = 'new/app/router/Route.go'; Start = 1; End = 37 },
    @{ Path = 'new/app/router/webApi2/webApi.go'; Start = 1; End = 71 },
    @{ Path = 'new/app/controller/admin/KuaiYanFull.go'; Start = 1; End = 530 },
    @{ Path = 'new/app/logic/common/VMP/VMP.go'; Start = 1; End = 186 },
    @{ Path = 'new/app/logic/common/rmbPay/rmbPay.go'; Start = 1; End = 947 },
    @{ Path = 'new/app/logic/common/ka/ka.go'; Start = 1; End = 837 }
)

function Escape-Html {
    param([string]$Text)
    if ($null -eq $Text) { return '' }
    return $Text.Replace('&', '&amp;').Replace('<', '&lt;').Replace('>', '&gt;').Replace('"', '&quot;')
}

function Ensure-Utf8Bom {
    param([string]$Path)
    $utf8NoBom = New-Object System.Text.UTF8Encoding($false)
    $utf8Bom = New-Object System.Text.UTF8Encoding($true)
    $content = [System.IO.File]::ReadAllText($Path, $utf8NoBom)
    [System.IO.File]::WriteAllText($Path, $content, $utf8Bom)
}

function Save-WordDocument {
    param(
        [object]$Document,
        [string]$DocxPath,
        [string]$PdfPath
    )

    try {
        $Document.SaveAs2([ref]$DocxPath, [ref]16)
    }
    catch {
        $Document.SaveAs([ref]$DocxPath, [ref]16)
    }
    $Document.ExportAsFixedFormat($PdfPath, 17)
}

function Normalize-DocumentImages {
    param([object]$Document)

    $maxWidthPoints = 400

    foreach ($shape in @($Document.InlineShapes)) {
        try {
            $shape.LockAspectRatio = -1
            if ($shape.Width -gt $maxWidthPoints) {
                $shape.Width = $maxWidthPoints
            }
            $shape.Range.ParagraphFormat.Alignment = 1
        }
        catch {
        }
    }

    foreach ($shape in @($Document.Shapes)) {
        try {
            $shape.LockAspectRatio = -1
            if ($shape.Width -gt $maxWidthPoints) {
                $shape.Width = $maxWidthPoints
            }
        }
        catch {
        }
    }
}

function Convert-HtmlToOffice {
    param(
        [string]$HtmlPath,
        [string]$DocxPath,
        [string]$PdfPath,
        [switch]$NormalizeImages
    )

    $word = New-Object -ComObject Word.Application
    $word.Visible = $false
    try {
        $document = $word.Documents.Open($HtmlPath)
        if ($NormalizeImages) {
            Normalize-DocumentImages -Document $document
        }
        Save-WordDocument -Document $document -DocxPath $DocxPath -PdfPath $PdfPath
        $document.Close()
    }
    finally {
        $word.Quit()
        [System.Runtime.InteropServices.Marshal]::ReleaseComObject($word) | Out-Null
        [GC]::Collect()
        [GC]::WaitForPendingFinalizers()
    }
}

$allLines = New-Object System.Collections.Generic.List[object]
foreach ($selection in $sourceSelections) {
    $fullPath = Join-Path $root $selection.Path
    $content = Get-Content -LiteralPath $fullPath -Encoding UTF8
    for ($i = $selection.Start; $i -le $selection.End; $i++) {
        $allLines.Add([pscustomobject]@{
            File = $selection.Path
            Line = $i
            Text = $content[$i - 1]
        })
    }
}

if ($allLines.Count -ne 3000) {
    throw ("Source line count must be 3000, got {0}." -f $allLines.Count)
}

$sourceHtmlPath = Join-Path $outDir 'FNKY_SourceMaterial_V1.0.449.html'
$sourceDocxPath = Join-Path $outDir 'FNKY_SourceMaterial_V1.0.449.docx'
$sourcePdfPath = Join-Path $outDir 'FNKY_SourceMaterial_V1.0.449.pdf'

$inputHtmls = Get-ChildItem -LiteralPath $outDir -File -Filter '*.html' |
    Where-Object { $_.FullName -ne $sourceHtmlPath } |
    Sort-Object Length -Descending

if ($inputHtmls.Count -ne 2) {
    throw ("Expected exactly two input html files, got {0}." -f $inputHtmls.Count)
}

$designHtml = $inputHtmls[0]
$noteHtml = $inputHtmls[1]

$designDocxPath = [System.IO.Path]::ChangeExtension($designHtml.FullName, '.docx')
$designPdfPath = [System.IO.Path]::ChangeExtension($designHtml.FullName, '.pdf')
$noteDocxPath = [System.IO.Path]::ChangeExtension($noteHtml.FullName, '.docx')
$notePdfPath = [System.IO.Path]::ChangeExtension($noteHtml.FullName, '.pdf')

$fileListRows = foreach ($selection in $sourceSelections) {
    "<tr><td>$(Escape-Html $selection.Path)</td><td>$($selection.Start)</td><td>$($selection.End)</td><td>$($selection.End - $selection.Start + 1)</td></tr>"
}

$html = New-Object System.Text.StringBuilder
$null = $html.AppendLine('<!DOCTYPE html>')
$null = $html.AppendLine('<html lang="zh-CN"><head><meta charset="UTF-8" /><meta http-equiv="Content-Type" content="text/html; charset=utf-8" />')
$null = $html.AppendLine('<title>&#36719;&#20214;&#33879;&#20316;&#26435;&#28304;&#31243;&#24207;&#37492;&#21035;&#26448;&#26009; V1.0.449</title>')
$null = $html.AppendLine('<style>')
$null = $html.AppendLine('@page { size: A4; margin: 10mm 10mm 12mm 10mm; }')
$null = $html.AppendLine('body { font-family: "Microsoft YaHei", "SimSun", sans-serif; margin: 0; color: #111; }')
$null = $html.AppendLine('.cover { page-break-after: always; min-height: 250mm; }')
$null = $html.AppendLine('.cover h1 { text-align: center; font-size: 20pt; margin: 70mm 0 8mm 0; }')
$null = $html.AppendLine('.cover h2 { text-align: center; font-size: 14pt; margin: 0 0 18mm 0; font-weight: normal; }')
$null = $html.AppendLine('.cover table { width: 100%; border-collapse: collapse; font-size: 10.5pt; }')
$null = $html.AppendLine('.cover td, .cover th { border: 1px solid #444; padding: 7px 8px; }')
$null = $html.AppendLine('.cover p { font-size: 10.5pt; line-height: 1.7; margin: 6pt 0; }')
$null = $html.AppendLine('.page { page-break-after: always; position: relative; min-height: 270mm; }')
$null = $html.AppendLine('.code { width: 100%; border-collapse: collapse; table-layout: fixed; font-family: "Consolas", "Courier New", monospace; font-size: 8.5pt; line-height: 1.1; }')
$null = $html.AppendLine('.code td { border: none; padding: 0.55mm 0; vertical-align: top; }')
$null = $html.AppendLine('.id { width: 32mm; color: #333; padding-right: 2mm; white-space: nowrap; }')
$null = $html.AppendLine('.text { white-space: pre; overflow: hidden; }')
$null = $html.AppendLine('.footer { position: absolute; bottom: 0; left: 0; right: 0; text-align: center; font-size: 9pt; color: #333; }')
$null = $html.AppendLine('</style></head><body>')
$null = $html.AppendLine('<section class="cover">')
$null = $html.AppendLine('<h1>&#36719;&#20214;&#33879;&#20316;&#26435;&#28304;&#31243;&#24207;&#37492;&#21035;&#26448;&#26009;</h1>')
$null = $html.AppendLine('<h2>&#39134;&#40479;&#24555;&#39564;&#32593;&#32476;&#31995;&#32479; V1.0.449</h2>')
$null = $html.AppendLine('<p>&#35828;&#26126;&#65306;&#26412;&#26448;&#26009;&#25353; 60 &#39029;&#12289;&#27599;&#39029; 50 &#34892;&#29983;&#25104;&#12290;&#21069; 10 &#39029;&#20026;&#31243;&#24207;&#21069;&#37096;&#32467;&#26500;&#19982; WebApi &#20837;&#21475;&#20195;&#30721;&#65292;&#21518;&#32493; 50 &#39029;&#20026;&#36830;&#32493;&#36873;&#21462;&#30340;&#20195;&#34920;&#24615;&#26680;&#24515;&#19994;&#21153;&#20195;&#30721;&#12290;</p>')
$null = $html.AppendLine('<table><tr><th>&#25991;&#20214;&#36335;&#24452;</th><th>&#36215;&#22987;&#34892;</th><th>&#32467;&#26463;&#34892;</th><th>&#34892;&#25968;</th></tr>')
$null = $html.AppendLine(($fileListRows -join [Environment]::NewLine))
$null = $html.AppendLine('</table></section>')

for ($pageIndex = 0; $pageIndex -lt 60; $pageIndex++) {
    $start = $pageIndex * 50
    $end = $start + 49
    $null = $html.AppendLine('<section class="page"><table class="code">')
    for ($i = $start; $i -le $end; $i++) {
        $line = $allLines[$i]
        $fileName = [System.IO.Path]::GetFileName($line.File)
        $id = '{0}:{1}' -f $fileName, $line.Line
        $text = Escape-Html $line.Text
        $null = $html.AppendLine("<tr><td class='id'>$id</td><td class='text'>$text</td></tr>")
    }
    $pageNo = $pageIndex + 1
    $null = $html.AppendLine("</table><div class='footer'>&#31532; $pageNo &#39029;</div></section>")
}

$null = $html.AppendLine('</body></html>')
$html.ToString() | Set-Content -LiteralPath $sourceHtmlPath -Encoding UTF8

Ensure-Utf8Bom -Path $sourceHtmlPath
Ensure-Utf8Bom -Path $designHtml.FullName
Ensure-Utf8Bom -Path $noteHtml.FullName

Convert-HtmlToOffice -HtmlPath $sourceHtmlPath -DocxPath $sourceDocxPath -PdfPath $sourcePdfPath
Convert-HtmlToOffice -HtmlPath $designHtml.FullName -DocxPath $designDocxPath -PdfPath $designPdfPath -NormalizeImages
Convert-HtmlToOffice -HtmlPath $noteHtml.FullName -DocxPath $noteDocxPath -PdfPath $notePdfPath

Write-Output 'Generated files:'
Write-Output $sourceDocxPath
Write-Output $sourcePdfPath
Write-Output $designDocxPath
Write-Output $designPdfPath
Write-Output $noteDocxPath
Write-Output $notePdfPath
