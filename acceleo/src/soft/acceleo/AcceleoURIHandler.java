package soft.acceleo;

import java.util.Arrays;
import java.util.Stack;

import org.eclipse.acceleo.internal.parser.compiler.IAcceleoParserURIHandler;
import org.eclipse.emf.common.util.URI;

public class AcceleoURIHandler implements IAcceleoParserURIHandler {

	private static final String SEGMENT_EMPTY = "";
	private static final String SEGMENT_SELF = ".";
	private static final String SEGMENT_PARENT = "..";
	
	private String[] resolve( String[] segments ) {
		Stack<String> result = new Stack<String>();
		for ( String segment : segments ) {
			if ( SEGMENT_PARENT.equals(segment) ) {
				result.pop();
			}
			else if ( SEGMENT_SELF.equals(segment) ) {
				
			}
			else
				result.push(segment);
		}
		return (String[]) result.toArray(new String[result.size()]);
	}
	
	@Override
	public URI transform(URI uri) {
		URI newURI = uri;
		String[] segments = uri.segments();
		if ( Arrays.stream(segments).anyMatch(s -> SEGMENT_EMPTY.equals(s) || SEGMENT_SELF.equals(s) || SEGMENT_PARENT.equals(s) )) {
			newURI = URI.createHierarchicalURI( uri.scheme(), uri.authority(), uri.device(), resolve(segments), uri.query(), uri.fragment());
		}
		newURI = toAcceleoJar( newURI );
		return newURI;
	}

	
	private URI toAcceleoJar( URI uri ) {
		String uriStr = uri.toString();
		if (uriStr.startsWith("jar:file:")) {
			int indexOf = uriStr.indexOf(".jar!/");
			if (indexOf != -1) {
				String name = uriStr;
				name = name.substring(0, indexOf);
				name = name.substring( name.lastIndexOf('/') + 1);
				name = "acceleo-jar:" + name + ".jar!/" + uriStr.substring(indexOf + ".jar!/".length());
				uri = URI.createURI(name);
			}
		}
		return uri;
	}
	
}
