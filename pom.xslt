<xsl:stylesheet version="3.0" 
    xmlns="http://maven.apache.org/POM/4.0.0" 
    xmlns:pom="http://maven.apache.org/POM/4.0.0" 
    xmlns:xsl="http://www.w3.org/1999/XSL/Transform">
    
    <xsl:output method="xml" omit-xml-declaration="yes" indent="yes"/>
    <xsl:strip-space elements="*"/>
    
    <xsl:param name="artifactId"/>
    <xsl:param name="version"/>
    
    <xsl:template match="/">
        <xsl:variable name="inputs" 
            select="collection(iri-to-uri('.?select=pom.xml;recurse=yes'))" 
            as="document-node()*"/>
        <xsl:for-each select="$inputs">
          <xsl:variable name="uri" select="document-uri()"/>
          <xsl:result-document href="{$uri}">
            <xsl:apply-templates mode="inPom" select="@*|node()"/>
          </xsl:result-document>
        </xsl:for-each>
    </xsl:template>
    
    <!--identity template copies everything forward by default-->
    <xsl:template match="node()|@*" mode="inPom">
        <xsl:copy>
            <xsl:apply-templates mode="inPom" select="@*|node()"/>
        </xsl:copy>
    </xsl:template>
    
    <xsl:template match="pom:version/text()" mode="inPom">
        <xsl:choose>
            <xsl:when test="../../pom:artifactId = $artifactId">
                <xsl:value-of select="$version"/>
            </xsl:when>
            <xsl:otherwise>
                <xsl:copy-of select="."/>
            </xsl:otherwise>
        </xsl:choose>
    </xsl:template>
    
</xsl:stylesheet>
